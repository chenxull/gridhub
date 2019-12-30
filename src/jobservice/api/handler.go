package api

import (
	"encoding/json"
	"fmt"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/query"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/utils"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/core"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/errs"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/job"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	totalHeaderKey = "Total-Count"
	nextCursorKey  = "Next-Cursor"
)

// Handler defines approaches to handle the http requests.
type Handler interface {
	HandlerLaunchJobReq(w http.ResponseWriter, req *http.Request)
	// HandleGetJobReq is used to handle the job stats query request.
	HandleGetJobReq(w http.ResponseWriter, req *http.Request)

	// HandleJobActionReq is used to handle the job action requests (stop/retry).
	HandleJobActionReq(w http.ResponseWriter, req *http.Request)

	// HandleCheckStatusReq is used to handle the job service healthy status checking request.
	HandleCheckStatusReq(w http.ResponseWriter, req *http.Request)

	// HandleJobLogReq is used to handle the request of getting job logs
	HandleJobLogReq(w http.ResponseWriter, req *http.Request)

	// HandleJobLogReq is used to handle the request of getting periodic executions
	HandlePeriodicExecutions(w http.ResponseWriter, req *http.Request)

	// HandleGetJobsReq is used to handle the request of getting jobs
	HandleGetJobsReq(w http.ResponseWriter, req *http.Request)
}

func writeDate(w http.ResponseWriter, byte []byte) {
	if _, err := w.Write(byte); err != nil {
		logger.Errorf("writer write error:%s", err)
	}
}

// DefaultHandler is the default request handler which implements the Handler interface.
type DefaultHandler struct {
	controller core.Interface
}

// NewDefaultHandler is constructor of DefaultHandler.
func NewDefaultHandler(ctl core.Interface) *DefaultHandler {
	return &DefaultHandler{
		controller: ctl,
	}
}

func (dh *DefaultHandler) HandlerLaunchJobReq(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		dh.handleError(w, req, http.StatusInternalServerError, errs.ReadRequestBodyError(err))
		return
	}

	//unmarshal data
	jobReq := &job.Request{}
	if err = json.Unmarshal(data, jobReq); err != nil {
		dh.handleError(w, req, http.StatusInternalServerError, errs.HandleJSONDataError(err))
		return
	}
	// Pass request to the controller for the follow-up.
	jobStats, err := dh.controller.LaunchJob(jobReq)

	if err != nil {
		code := http.StatusInternalServerError
		if errs.IsBadRequestError(err) {
			code = http.StatusBadRequest
		} else if errs.IsConflictError(err) {
			code = http.StatusConflict
		} else {
			err = errs.LaunchJobError(err)
		}
		dh.handleError(w, req, code, err)
		return
	}
	dh.handleJSONData(w, req, http.StatusAccepted, jobStats)

}

func (dh *DefaultHandler) HandleGetJobReq(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	jobID := vars["job_id"]
	jobStats, err := dh.controller.GetJob(jobID)
	if err != nil {
		code := http.StatusInternalServerError
		if errs.IsObjectNotFoundError(err) {
			code = http.StatusNotFound
		} else if errs.IsBadRequestError(err) {
			code = http.StatusBadRequest
		} else {
			err = errs.GetJobStatsError(err)
		}
		dh.handleError(w, req, code, err)
		return
	}
	dh.handleJSONData(w, req, http.StatusOK, jobStats)

}

func (dh *DefaultHandler) HandleJobActionReq(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	jobID := vars["job_id"]

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		dh.handleError(w, req, http.StatusInternalServerError, errs.ReadRequestBodyError(err))
		return
	}

	// unmarshal data
	jobActionReq := &job.ActionRequest{}
	if err = json.Unmarshal(data, jobActionReq); err != nil {
		dh.handleError(w, req, http.StatusInternalServerError, errs.HandleJSONDataError(err))
		return
	}

	// Only support stop command now
	cmd := job.OPCommand(jobActionReq.Action)
	if !cmd.IsStop() {
		dh.handleError(w, req, http.StatusNotImplemented, errs.UnknownActionNameError(errors.Errorf("command: %s", jobActionReq.Action)))
		return
	}

	// Stop job
	if err := dh.controller.StopJob(jobID); err != nil {
		code := http.StatusInternalServerError
		if errs.IsObjectNotFoundError(err) {
			code = http.StatusNotFound
		} else if errs.IsBadRequestError(err) {
			code = http.StatusBadRequest
		} else {
			err = errs.StopJobError(err)
		}
		dh.handleError(w, req, code, err)
		return
	}

	dh.log(req, http.StatusNoContent, string(data))

	w.WriteHeader(http.StatusNoContent) // only header, no content returned
}

func (dh *DefaultHandler) HandleCheckStatusReq(w http.ResponseWriter, req *http.Request) {
	stats, err := dh.controller.CheckStatus()
	if err != nil {
		dh.handleError(w, req, http.StatusInternalServerError, errs.CheckStatsError(err))
		return
	}

	dh.handleJSONData(w, req, http.StatusOK, stats)
}

func (dh *DefaultHandler) HandleJobLogReq(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	jobID := vars["job_id"]

	if strings.Contains(jobID, "..") || strings.ContainsRune(jobID, os.PathSeparator) {
		dh.handleError(w, req, http.StatusBadRequest, errors.Errorf("invalid Job ID: %s", jobID))
		return
	}

	logData, err := dh.controller.GetJobLogData(jobID)
	if err != nil {
		code := http.StatusInternalServerError
		if errs.IsObjectNotFoundError(err) {
			code = http.StatusNotFound
		} else if errs.IsBadRequestError(err) {
			code = http.StatusBadRequest
		} else {
			err = errs.GetJobLogError(err)
		}
		dh.handleError(w, req, code, err)
		return
	}

	dh.log(req, http.StatusOK, "")

	w.WriteHeader(http.StatusOK)
	writeDate(w, logData)
}

func (dh *DefaultHandler) HandlePeriodicExecutions(w http.ResponseWriter, req *http.Request) {
	// Get param
	vars := mux.Vars(req)
	jobID := vars["job_id"]

	// Get query params
	q := extractQuery(req)

	executions, total, err := dh.controller.GetPeriodicExecutions(jobID, q)
	if err != nil {
		code := http.StatusInternalServerError
		if errs.IsObjectNotFoundError(err) {
			code = http.StatusNotFound
		} else if errs.IsBadRequestError(err) {
			code = http.StatusBadRequest
		} else {
			err = errs.GetPeriodicExecutionError(err)
		}
		dh.handleError(w, req, code, err)
		return
	}

	w.Header().Add(totalHeaderKey, fmt.Sprintf("%d", total))
	dh.handleJSONData(w, req, http.StatusOK, executions)
}

func (dh *DefaultHandler) log(req *http.Request, code int, text string) {
	logger.Debugf("Serve http request '%s %s': %d %s", req.Method, req.URL.String(), code, text)
}
func (dh *DefaultHandler) handleError(w http.ResponseWriter, req *http.Request, code int, err error) {
	// Log all errors
	logger.Errorf("Serve http request '%s %s' error: %d %s", req.Method, req.URL.String(), code, err.Error())

	w.WriteHeader(code)
	writeDate(w, []byte(err.Error()))
}

func (dh *DefaultHandler) handleJSONData(w http.ResponseWriter, req *http.Request, code int, object interface{}) {
	data, err := json.Marshal(object)
	if err != nil {
		dh.handleError(w, req, http.StatusInternalServerError, errs.HandleJSONDataError(err))
		return
	}
	logger.Debugf("Serve http request '%s %s': %d %s", req.Method, req.URL.String(), code, data)
	w.Header().Set(http.CanonicalHeaderKey("Accept"), "application/json")
	w.Header().Set(http.CanonicalHeaderKey("content-type"), "application/json")
	w.WriteHeader(code)
	writeDate(w, data)

}

func extractQuery(req *http.Request) *query.Parameter {
	q := &query.Parameter{
		PageNumber: 1,
		PageSize:   query.DefaultPageSize,
		Extras:     make(query.ExtraParameters),
	}

	queries := req.URL.Query()
	// Page number
	p := queries.Get(query.ParamKeyPage)
	if !utils.IsEmptyStr(p) {
		if pv, err := strconv.ParseUint(p, 10, 32); err == nil {
			if pv > 1 {
				q.PageNumber = uint(pv)
			}
		}
	}

	// Page number
	size := queries.Get(query.ParamKeyPageSize)
	if !utils.IsEmptyStr(size) {
		if pz, err := strconv.ParseUint(size, 10, 32); err == nil {
			if pz > 0 {
				q.PageSize = uint(pz)
			}
		}
	}

	// Extra query parameters
	nonStoppedOnly := queries.Get(query.ParamKeyNonStoppedOnly)
	if !utils.IsEmptyStr(nonStoppedOnly) {
		if nonStoppedOnlyV, err := strconv.ParseBool(nonStoppedOnly); err == nil {
			q.Extras.Set(query.ExtraParamKeyNonStoppedOnly, nonStoppedOnlyV)
		}
	}

	// Extra job kind query param
	jobKind := queries.Get(query.ParamKeyJobKind)
	if !utils.IsEmptyStr(jobKind) {
		q.Extras.Set(query.ExtraParamKeyKind, jobKind)
	}

	// Extra query cursor
	cursorV := queries.Get(query.ParamKeyCursor)
	if !utils.IsEmptyStr(cursorV) {
		if cursor, err := strconv.ParseInt(cursorV, 10, 32); err == nil {
			q.Extras.Set(query.ExtraParamKeyCursor, cursor)
		}
	}

	return q
}
func (dh *DefaultHandler) HandleGetJobsReq(w http.ResponseWriter, req *http.Request) {
	panic("implement me")
}
