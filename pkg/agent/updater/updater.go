package updater

import (
	"context"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/deviceplane/deviceplane/pkg/agent/utils"
	"github.com/deviceplane/deviceplane/pkg/engine"
	"github.com/deviceplane/deviceplane/pkg/models"
)

type Updater struct {
	engine    engine.Engine
	projectID string
	version   string

	desiredSpec models.Service
	once        sync.Once
	lock        sync.RWMutex
}

func NewUpdater(engine engine.Engine, projectID, version string) *Updater {
	return &Updater{
		engine:    engine,
		projectID: projectID,
		version:   version,
	}
}

func (u *Updater) SetDesiredSpec(desiredSpec models.Service) {
	u.lock.Lock()
	u.desiredSpec = desiredSpec
	u.lock.Unlock()

	u.once.Do(func() {
		go u.updater()
	})
}

func (u *Updater) updater() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		u.lock.RLock()
		desiredSpec := u.desiredSpec
		u.lock.RUnlock()

		var desiredVersion string
		if parts := strings.Split(desiredSpec.Image, ":"); len(parts) == 2 {
			desiredVersion = parts[1]
		} else {
			log.Error("invalid agent image")
			goto cont
		}

		if desiredVersion != "" && desiredVersion != u.version {
			u.update(desiredSpec, desiredVersion)
		}

	cont:
		select {
		case <-ticker.C:
			continue
		}
	}
}

func (u *Updater) update(desiredSpec models.Service, desiredVersion string) {
	instances := utils.ContainerList(context.TODO(), u.engine, nil, map[string]string{
		models.AgentVersionLabel: desiredVersion,
	}, true)

	if len(instances) > 0 {
		return
	}

	utils.ImagePull(context.TODO(), u.engine, desiredSpec.Image, ioutil.Discard)

	instanceID := utils.ContainerCreate(
		context.TODO(),
		u.engine,
		"",
		withCommandInterpolation(withAgentVersionLabel(desiredSpec, desiredVersion), u.projectID),
	)

	utils.ContainerStart(context.TODO(), u.engine, instanceID)
}

func withAgentVersionLabel(s models.Service, version string) models.Service {
	if s.Labels == nil {
		s.Labels = make(map[string]string)
	}
	s.Labels[models.AgentVersionLabel] = version
	return s
}

func withCommandInterpolation(s models.Service, projectID string) models.Service {
	var command []string
	for _, arg := range s.Command {
		command = append(command, strings.ReplaceAll(arg, "$PROJECT", projectID))
	}
	s.Command = command
	return s
}
