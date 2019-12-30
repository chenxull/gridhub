package api

import (
	"fmt"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/errs"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"net/http"
)

const (
	baseRoute  = "/api"
	apiVersion = "/v1"
)

//Router defines the related routes for the job service and directs the request to the right handler
//method

type Router interface {
	//ServerHTTP used to handle the http requests
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type BaseRouter struct {
	// Use mux to keep the routes mapping.
	router *mux.Router

	//Handler used to handle the request
	handler Handler

	//Do Auth
	authenticator Authenticator
}

//NewBaseRouter is the constructor of BaseRouter
func NewBaseRouter(handler Handler, authenticator Authenticator) Router {
	br := &BaseRouter{
		router:        mux.NewRouter(),
		handler:       handler,
		authenticator: authenticator,
	}

	//Register routes here
	br.registerRoutes()
	return br
}

// ServeHTTP is the implementation of Router interface.
func (br *BaseRouter) ServerHTTP(w http.ResponseWriter, req *http.Request) {
	// No auth required for /stats as it is a health check endpoint
	// Do auth for other services
	if req.URL.String() != fmt.Sprintf("%s/%s/stats", baseRoute, apiVersion) {
		if err := br.authenticator.DoAuth(req); err != nil {
			authErr := errs.UnauthorizedError(err)
			if authErr == nil {
				authErr = errors.Errorf("unauthorized: %s", err)
			}
			logger.Errorf("Serve http request '%s %s' failed with error: %s", req.Method, req.URL.String(), authErr.Error())
			w.WriteHeader(http.StatusUnauthorized)
			writeDate(w, []byte(authErr.Error()))

			return
		}
	}

	// Directly pass requests to the server mux
	br.router.ServeHTTP(w, req)
}

func (br *BaseRouter) registerRoutes() {
	// remove the prefix of of the request router
	subRouter := br.router.PathPrefix(fmt.Sprintf("%s/%s", baseRoute, apiVersion)).Subrouter()

	subRouter.HandleFunc("/jobs", br.handler.HandlerLaunchJobReq).Methods(http.MethodPost)
	subRouter.HandleFunc("/jobs", br.handler.HandleGetJobsReq).Methods(http.MethodGet)
	subRouter.HandleFunc("/jobs/{job_id}", br.handler.HandleGetJobReq).Methods(http.MethodGet)
	subRouter.HandleFunc("/jobs/{job_id}", br.handler.HandleJobActionReq).Methods(http.MethodPost)
	subRouter.HandleFunc("/jobs/{job_id}/log", br.handler.HandleJobLogReq).Methods(http.MethodGet)
	subRouter.HandleFunc("/stats", br.handler.HandleCheckStatusReq).Methods(http.MethodGet)
	subRouter.HandleFunc("/jobs/{job_id}/executions", br.handler.HandlePeriodicExecutions).Methods(http.MethodGet)

}
