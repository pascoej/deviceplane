package service

import (
	"github.com/deviceplane/deviceplane/pkg/codes"
	"github.com/deviceplane/deviceplane/pkg/utils"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Service) logs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["application"]
	service := vars["service"]

	resp, err := s.supervisorLookup.GetServiceLogs(r.Context(), applicationID, service)
	if err != nil {
		w.WriteHeader(codes.StatusServiceLogsNotAvailable)
		return
	}

	utils.Respond(w, resp)
}
