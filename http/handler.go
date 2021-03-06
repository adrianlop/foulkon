package http

import (
	"encoding/json"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/satori/go.uuid"
	"github.com/tecsisa/foulkon/api"
	"github.com/tecsisa/foulkon/foulkon"
)

const (
	// Constants for values in url
	USER_ID     = "userid"
	GROUP_NAME  = "groupname"
	POLICY_NAME = "policyname"
	ORG_NAME    = "orgname"

	// URI Path param prefix
	URI_PATH_PREFIX = "/:"

	// API root reference
	API_ROOT      = "/api"
	API_VERSION_1 = API_ROOT + "/v1"

	// Organization API ROOT
	ORG_ROOT = "/organizations/:" + ORG_NAME

	// User API urls
	USER_ROOT_URL      = API_VERSION_1 + "/users"
	USER_ID_URL        = USER_ROOT_URL + URI_PATH_PREFIX + USER_ID
	USER_ID_GROUPS_URL = USER_ID_URL + "/groups"

	// Group organization API urls
	GROUP_ORG_ROOT_URL       = API_VERSION_1 + ORG_ROOT + "/groups"
	GROUP_ID_URL             = GROUP_ORG_ROOT_URL + URI_PATH_PREFIX + GROUP_NAME
	GROUP_ID_USERS_URL       = GROUP_ID_URL + "/users"
	GROUP_ID_USERS_ID_URL    = GROUP_ID_USERS_URL + URI_PATH_PREFIX + USER_ID
	GROUP_ID_POLICIES_URL    = GROUP_ID_URL + "/policies"
	GROUP_ID_POLICIES_ID_URL = GROUP_ID_POLICIES_URL + URI_PATH_PREFIX + POLICY_NAME

	// Policy API urls
	POLICY_ROOT_URL      = API_VERSION_1 + ORG_ROOT + "/policies"
	POLICY_ID_URL        = POLICY_ROOT_URL + URI_PATH_PREFIX + POLICY_NAME
	POLICY_ID_GROUPS_URL = POLICY_ROOT_URL + URI_PATH_PREFIX + POLICY_NAME + "/groups"

	// Authorization URLs
	RESOURCE_URL = API_VERSION_1 + "/resource"

	// HTTP Header
	REQUEST_ID_HEADER = "Request-ID"
)

// WORKER

type WorkerHandler struct {
	worker *foulkon.Worker
}

func (a *WorkerHandler) TransactionLog(r *http.Request, requestID string, userID string, msg string) {

	// TODO: X-Forwarded headers?
	//for header, _ := range r.Header {
	//	println(header, ": ", r.Header.Get(header))
	//}

	a.worker.Logger.WithFields(logrus.Fields{
		"requestID": requestID,
		"method":    r.Method,
		"URI":       r.RequestURI,
		"address":   r.RemoteAddr,
		"user":      userID,
	}).Info(msg)
}

// Handler returns http.Handler for the APIs.
func WorkerHandlerRouter(worker *foulkon.Worker) http.Handler {
	// Create the muxer to handle the actual endpoints
	router := httprouter.New()

	workerHandler := WorkerHandler{worker: worker}

	// User api
	router.GET(USER_ROOT_URL, workerHandler.HandleListUsers)
	router.POST(USER_ROOT_URL, workerHandler.HandleAddUser)

	router.GET(USER_ID_URL, workerHandler.HandleGetUserByExternalID)
	router.PUT(USER_ID_URL, workerHandler.HandleUpdateUser)
	router.DELETE(USER_ID_URL, workerHandler.HandleRemoveUser)

	router.GET(USER_ID_GROUPS_URL, workerHandler.HandleListGroupsByUser)

	// Group api
	router.POST(GROUP_ORG_ROOT_URL, workerHandler.HandleAddGroup)
	router.GET(GROUP_ORG_ROOT_URL, workerHandler.HandleListGroups)

	router.DELETE(GROUP_ID_URL, workerHandler.HandleRemoveGroup)
	router.GET(GROUP_ID_URL, workerHandler.HandleGetGroupByName)
	router.PUT(GROUP_ID_URL, workerHandler.HandleUpdateGroup)

	router.GET(GROUP_ID_USERS_URL, workerHandler.HandleListMembers)

	router.POST(GROUP_ID_USERS_ID_URL, workerHandler.HandleAddMember)
	router.DELETE(GROUP_ID_USERS_ID_URL, workerHandler.HandleRemoveMember)

	router.GET(GROUP_ID_POLICIES_URL, workerHandler.HandleListAttachedGroupPolicies)

	router.POST(GROUP_ID_POLICIES_ID_URL, workerHandler.HandleAttachPolicyToGroup)
	router.DELETE(GROUP_ID_POLICIES_ID_URL, workerHandler.HandleDetachPolicyToGroup)

	// Special endpoint without organization URI for groups
	router.GET(API_VERSION_1+"/groups", workerHandler.HandleListAllGroups)

	// Policy api
	router.GET(POLICY_ROOT_URL, workerHandler.HandleListPolicies)
	router.POST(POLICY_ROOT_URL, workerHandler.HandleAddPolicy)

	router.DELETE(POLICY_ID_URL, workerHandler.HandleRemovePolicy)

	router.GET(POLICY_ID_URL, workerHandler.HandleGetPolicyByName)
	router.PUT(POLICY_ID_URL, workerHandler.HandleUpdatePolicy)

	router.GET(POLICY_ID_GROUPS_URL, workerHandler.HandleListAttachedGroups)

	// Special endpoint without organization URI for policies
	router.GET(API_VERSION_1+"/policies", workerHandler.HandleListAllPolicies)

	// Resources authorized endpoint
	router.POST(RESOURCE_URL, workerHandler.HandleGetAuthorizedExternalResources)

	// Return handler with request logging
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewV4().String()
		r.Header.Set(REQUEST_ID_HEADER, requestID)
		w.Header().Add(REQUEST_ID_HEADER, requestID)
		worker.Authenticator.Authenticate(router).ServeHTTP(w, r)
		userID, _ := worker.Authenticator.GetAuthenticatedUser(r)
		workerHandler.TransactionLog(r, requestID, userID, "")
	})
}

// HTTP WORKER responses

// 2xx RESPONSES

func (a *WorkerHandler) RespondOk(r *http.Request, requestInfo api.RequestInfo, w http.ResponseWriter, value interface{}) {
	b, err := json.Marshal(value)
	if err != nil {
		a.RespondInternalServerError(r, requestInfo, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (a *WorkerHandler) RespondCreated(r *http.Request, requestInfo api.RequestInfo, w http.ResponseWriter, value interface{}) {
	b, err := json.Marshal(value)
	if err != nil {
		a.RespondInternalServerError(r, requestInfo, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(b)
}

func (a *WorkerHandler) RespondNoContent(r *http.Request, requestInfo api.RequestInfo, w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// 4xx RESPONSES

func (a *WorkerHandler) RespondNotFound(r *http.Request, requestInfo api.RequestInfo, w http.ResponseWriter, apiError *api.Error) {
	w, err := writeErrorWithStatus(w, apiError, http.StatusNotFound)
	if err != nil {
		a.RespondInternalServerError(r, requestInfo, w)
		return
	}
}

func (a *WorkerHandler) RespondBadRequest(r *http.Request, requestInfo api.RequestInfo, w http.ResponseWriter, apiError *api.Error) {
	w, err := writeErrorWithStatus(w, apiError, http.StatusBadRequest)
	if err != nil {
		a.RespondInternalServerError(r, requestInfo, w)
		return
	}
}

func (a *WorkerHandler) RespondConflict(r *http.Request, requestInfo api.RequestInfo, w http.ResponseWriter, apiError *api.Error) {
	w, err := writeErrorWithStatus(w, apiError, http.StatusConflict)
	if err != nil {
		a.RespondInternalServerError(r, requestInfo, w)
		return
	}
}

func (a *WorkerHandler) RespondForbidden(r *http.Request, requestInfo api.RequestInfo, w http.ResponseWriter, apiError *api.Error) {
	w, err := writeErrorWithStatus(w, apiError, http.StatusForbidden)
	if err != nil {
		a.RespondInternalServerError(r, requestInfo, w)
		return
	}
}

// 5xx RESPONSES

func (a *WorkerHandler) RespondInternalServerError(r *http.Request, requestInfo api.RequestInfo, w http.ResponseWriter) {
	a.worker.Logger.WithFields(logrus.Fields{
		"requestID": requestInfo.RequestID,
		"method":    r.Method,
		"URI":       r.RequestURI,
		"address":   r.RemoteAddr,
		"user":      requestInfo.Identifier,
		"status":    http.StatusInternalServerError,
	}).Error("Internal server error")
	w.WriteHeader(http.StatusInternalServerError)
}

// Worker Aux method

func (w *WorkerHandler) GetRequestInfo(r *http.Request) api.RequestInfo {
	userID, admin := w.worker.Authenticator.GetAuthenticatedUser(r)
	return api.RequestInfo{
		Identifier: userID,
		Admin:      admin,
		RequestID:  r.Header.Get(REQUEST_ID_HEADER),
	}
}

// PROXY

type ProxyHandler struct {
	proxy  *foulkon.Proxy
	client *http.Client
}

func (h *ProxyHandler) TransactionErrorLog(r *http.Request, requestID string, workerRequestID string, msg string) {

	// TODO: X-Forwarded headers
	//for header, _ := range r.Header {
	//	println(header, ": ", r.Header.Get(header))
	//}

	h.proxy.Logger.WithFields(logrus.Fields{
		"requestID":       requestID,
		"method":          r.Method,
		"URI":             r.URL.EscapedPath(),
		"address":         r.RemoteAddr,
		"workerRequestID": workerRequestID,
	}).Error(msg)
}

func (h *ProxyHandler) TransactionLog(r *http.Request, requestID string, workerRequestID string, msg string) {

	// TODO: X-Forwarded headers
	//for header, _ := range r.Header {
	//	println(header, ": ", r.Header.Get(header))
	//}

	h.proxy.Logger.WithFields(logrus.Fields{
		"requestID":       requestID,
		"method":          r.Method,
		"URI":             r.URL.EscapedPath(),
		"address":         r.RemoteAddr,
		"workerRequestID": workerRequestID,
	}).Info(msg)
}

func (h *ProxyHandler) RespondForbidden(w http.ResponseWriter, proxyErr *api.Error) {
	b, err := json.Marshal(proxyErr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write(b)
}

func (h *ProxyHandler) RespondBadRequest(w http.ResponseWriter, proxyErr *api.Error) {
	b, err := json.Marshal(proxyErr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(b)
}

func (h *ProxyHandler) RespondInternalServerError(w http.ResponseWriter, proxyErr *api.Error) {
	b, err := json.Marshal(proxyErr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(b)
}

// Handler returns an http.Handler for the Proxy including all resources defined in proxy file.
func ProxyHandlerRouter(proxy *foulkon.Proxy) http.Handler {
	// Create the muxer to handle the actual endpoints
	router := httprouter.New()

	proxyHandler := ProxyHandler{proxy: proxy, client: http.DefaultClient}

	for _, res := range proxy.APIResources {
		router.Handle(res.Method, res.Url, proxyHandler.HandleRequest(res))
	}

	return router
}

// Private Helper Methods
func writeErrorWithStatus(w http.ResponseWriter, apiError *api.Error, statusCode int) (http.ResponseWriter, error) {
	b, err := json.Marshal(apiError)
	if err != nil {
		return nil, err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(b)
	return w, nil
}
