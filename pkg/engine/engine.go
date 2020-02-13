package engine

import (
	"context"
	"errors"
	"io"

	"github.com/deviceplane/deviceplane/pkg/models"
)

var (
	ErrInstanceNotFound = errors.New("instance not found")
)

type Engine interface {
	CreateContainer(context.Context, string, models.Service) (string, error)
	InspectContainer(context.Context, string) (*InspectResponse, error)
	StartContainer(context.Context, string) error
	ListContainers(context.Context, map[string]struct{}, map[string]string, bool) ([]Instance, error)
	StopContainer(context.Context, string) error
	RemoveContainer(context.Context, string) error
	FetchContainerLogs(context.Context, string) (string, error)

	PullImage(context.Context, string, string, io.Writer) error
}

type Instance struct {
	ID      string
	Labels  map[string]string
	Running bool
}

type InspectResponse struct {
	PID int
}
