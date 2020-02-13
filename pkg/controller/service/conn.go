package service

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/deviceplane/deviceplane/pkg/agent/service/client"
	"github.com/deviceplane/deviceplane/pkg/codes"
	"github.com/deviceplane/deviceplane/pkg/models"
	"github.com/deviceplane/deviceplane/pkg/utils"
	"github.com/function61/holepunch-server/pkg/wsconnadapter"
	"github.com/gorilla/mux"
)

func (s *Service) initiateDeviceConnection(w http.ResponseWriter, r *http.Request, project models.Project, device models.Device) {
	s.withHijackedWebSocketConnection(w, r, func(conn net.Conn) {
		s.connman.Set(project.ID+device.ID, conn)
	})
}

var currentSSHCount int64

const currentSSHCountName = "internal.current_ssh_connection_count"

func (s *Service) initiateSSH(w http.ResponseWriter, r *http.Request,
	projectID, authenticatedUserID, authenticatedServiceAccountID,
	deviceID string,
) {
	s.withHijackedWebSocketConnection(w, r, func(conn net.Conn) {
		s.withDeviceConnection(w, r, projectID, deviceID, func(deviceConn net.Conn) {
			err := client.InitiateSSH(r.Context(), deviceConn)
			if err != nil {
				http.Error(w, err.Error(), codes.StatusDeviceConnectionFailure)
				return
			}

			sshCount := atomic.AddInt64(&currentSSHCount, 1)
			s.st.Gauge(currentSSHCountName, float64(sshCount), utils.InternalTags(projectID), 1)
			defer func() {
				sshCount := atomic.AddInt64(&currentSSHCount, -1)
				s.st.Gauge(currentSSHCountName, float64(sshCount), utils.InternalTags(projectID), 1)
			}()

			go io.Copy(deviceConn, conn)
			io.Copy(conn, deviceConn)
		})
	})
}

func (s *Service) initiateReboot(w http.ResponseWriter, r *http.Request,
	projectID, authenticatedUserID, authenticatedServiceAccountID,
	deviceID string,
) {
	s.withDeviceConnection(w, r, projectID, deviceID, func(deviceConn net.Conn) {
		resp, err := client.InitiateReboot(r.Context(), deviceConn)
		if err != nil {
			http.Error(w, err.Error(), codes.StatusDeviceConnectionFailure)
			return
		}

		utils.ProxyResponseFromDevice(w, resp)
	})
}

func (s *Service) deviceDebug(w http.ResponseWriter, r *http.Request,
	projectID, authenticatedUserID, authenticatedServiceAccountID,
	deviceID string,
) {
	path := r.URL.EscapedPath()
	dIndex := strings.Index(path, "/debug/")
	if dIndex == -1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	debugPath := path[dIndex:]

	s.withDeviceConnection(w, r, projectID, deviceID, func(deviceConn net.Conn) {
		req, err := http.NewRequestWithContext(
			r.Context(),
			r.Method,
			debugPath,
			r.Body,
		)
		if err != nil {
			http.Error(w, err.Error(), codes.StatusDeviceConnectionFailure)
			return
		}

		if err := req.Write(deviceConn); err != nil {
			http.Error(w, err.Error(), codes.StatusDeviceConnectionFailure)
			return
		}

		resp, err := http.ReadResponse(bufio.NewReader(deviceConn), req)
		if err != nil {
			http.Error(w, err.Error(), codes.StatusDeviceConnectionFailure)
			return
		}

		utils.ProxyResponseFromDevice(w, resp)
	})
}

func (s *Service) imagePullProgress(w http.ResponseWriter, r *http.Request,
	projectID, authenticatedUserID, authenticatedServiceAccountID,
	deviceID string,
) {
	vars := mux.Vars(r)
	applicationID := vars["application"]
	service := vars["service"]

	s.withDeviceConnection(w, r, projectID, deviceID, func(deviceConn net.Conn) {
		resp, err := client.GetImagePullProgress(r.Context(), deviceConn, applicationID, service)
		if err != nil {
			http.Error(w, err.Error(), codes.StatusDeviceConnectionFailure)
			return
		}

		utils.ProxyResponseFromDevice(w, resp)
	})
}

func (s *Service) hostMetrics(w http.ResponseWriter, r *http.Request,
	projectID, authenticatedUserID, authenticatedServiceAccountID,
	deviceID string,
) {
	s.withDeviceConnection(w, r, projectID, deviceID, func(deviceConn net.Conn) {
		resp, err := client.GetDeviceMetrics(r.Context(), deviceConn)
		if err != nil {
			http.Error(w, err.Error(), codes.StatusDeviceConnectionFailure)
			return
		}

		utils.ProxyResponseFromDevice(w, resp)
	})
}

func (s *Service) agentMetrics(w http.ResponseWriter, r *http.Request,
	projectID, authenticatedUserID, authenticatedServiceAccountID,
	deviceID string,
) {
	s.withDeviceConnection(w, r, projectID, deviceID, func(deviceConn net.Conn) {
		resp, err := client.GetAgentMetrics(r.Context(), deviceConn)
		if err != nil {
			http.Error(w, err.Error(), codes.StatusDeviceConnectionFailure)
			return
		}

		utils.ProxyResponseFromDevice(w, resp)
	})
}

func (s *Service) serviceMetrics(w http.ResponseWriter, r *http.Request,
	projectID, authenticatedUserID, authenticatedServiceAccountID,
	applicationID, deviceID string,
) {
	vars := mux.Vars(r)
	service := vars["service"]

	app, err := s.applications.GetApplication(r.Context(), applicationID, projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serviceMetricEndpointConfig, exists := app.MetricEndpointConfigs[service]
	if !exists {
		serviceMetricEndpointConfig.Port = models.DefaultMetricPort
		serviceMetricEndpointConfig.Path = models.DefaultMetricPath
	}

	s.withDeviceConnection(w, r, projectID, deviceID, func(deviceConn net.Conn) {
		resp, err := client.GetServiceMetrics(
			r.Context(), deviceConn, applicationID, service,
			serviceMetricEndpointConfig.Path, serviceMetricEndpointConfig.Port,
		)
		if err != nil {
			http.Error(w, err.Error(), codes.StatusDeviceConnectionFailure)
			return
		}

		utils.ProxyResponseFromDevice(w, resp)
	})
}

func (s *Service) serviceLogs(w http.ResponseWriter, r *http.Request,
	projectID, authenticatedUserID, authenticatedServiceAccountID,
	applicationID, deviceID string,
) {
	vars := mux.Vars(r)
	service := vars["service"]

	app, err := s.applications.GetApplication(r.Context(), applicationID, projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	serviceMetricEndpointConfig, exists := app.MetricEndpointConfigs[service]
	if !exists {
		serviceMetricEndpointConfig.Port = models.DefaultMetricPort
		serviceMetricEndpointConfig.Path = models.DefaultMetricPath
	}

	s.withDeviceConnection(w, r, projectID, deviceID, func(deviceConn net.Conn) {
		resp, err := client.GetServiceLogs(
			r.Context(), deviceConn, applicationID, service,
			serviceMetricEndpointConfig.Path, serviceMetricEndpointConfig.Port,
		)
		if err != nil {
			http.Error(w, err.Error(), codes.StatusDeviceConnectionFailure)
			return
		}

		utils.ProxyResponseFromDevice(w, resp)
	})
}

func (s *Service) withHijackedWebSocketConnection(w http.ResponseWriter, r *http.Request, f func(net.Conn)) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// We should set conn.CloseHandler() here

	f(wsconnadapter.New(conn))
}

func (s *Service) withDeviceConnection(w http.ResponseWriter, r *http.Request, projectID, deviceID string, f func(net.Conn)) {
	deviceConn, err := s.connman.Dial(r.Context(), projectID+deviceID)
	if err != nil {
		http.Error(w, err.Error(), codes.StatusDeviceConnectionFailure)
		return
	}
	defer deviceConn.Close()

	f(deviceConn)
}
