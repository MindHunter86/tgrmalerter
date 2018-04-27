package app

import "io/ioutil"
import "net/http"
import "encoding/json"
import "github.com/gorilla/mux"
import "github.com/gorilla/context"

import "github.com/satori/go.uuid"

// JSON response structs:
// Recomendations are taken from jsonapi.org:
type apiResponse struct {
	Data *responseData					`json:"data,omitempty"`
	Errors []*responseError			`json:"errors,omitempty"`
}
type responseData struct {
	Type string									`json:"type,omitempty"`
	Id int											`json:"id,omitempty"`
	Attributes *dataAttributes	`json:"attributes,omitempty"`
	Links *dataLinks						`json:"links,omitempty"`
}
type dataAttributes struct {
	Phone string								`json:"phone,omitempty"`
	Alert string								`json:"alert,omitempty"`
	Status string								`json:"status,omitempty"`
}
type dataLinks struct {
	Self string									`json:"self,omitempty"`
}
type responseError struct {
	Id string										`json:"id,omitempty"`
	Links *dataLinks						`json:"links,omitempty"`
	Code int										`json:"code,omitempty"`
	Status int									`json:"status,omitempty"`
	Title string								`json:"title,omitempty"`
	Detail string								`json:"detail,omitempty"`
	Source *errorSource					`json:"source,omitempty"`
}
type errorSource struct {
	Parameter string						`json:"parameter,omitempty"`
}

// JSON request structs:
type apiPostRequest struct {
	Data *requestData						`json:"data,omitempty"`
}
type requestData struct {
	Type string									`json:"type,omitempty"`
	Attributes *dataAttributes	`json:"attributes,omitempty"`
}

// request struct:
// TODO:
//  - check type and link
type httpRequest struct {
	// internal pointers:
	ap *api

	// payload:
	id string
}

func (m *httpRequest) create(a *api) *httpRequest {
	m.id = uuid.NewV4().String()
	m.ap = a
	return m
}

func (m *httpRequest) save(r *http.Request) error {
	m.ap.log.Debug().Msg("Saving request data")

	stmt,e := m.ap.sqldb.Prepare("INSERT INTO requests VALUES (?,?,?,?,?,?,?)"); if e != nil { return e }
	defer stmt.Close()

	if _,e = stmt.Exec(m.id, r.RemoteAddr, r.Method, r.ContentLength, r.RequestURI, 0, r.UserAgent()); e != nil { return e }

	return nil
}

func (m *httpRequest) saveStatus(status int) {
	stmt,e := m.ap.sqldb.Prepare("UPDATE requests SET status = ? WHERE id = ?"); if e != nil {
		m.ap.log.Error().Err(e).Msg("Could not prepare DB statement!"); return }
	defer stmt.Close()

	if _,e := stmt.Exec(status, m.id); e != nil {
		m.ap.log.Error().Err(e).Msg("Could not execute DB statement!"); return }
}


// internal Application struct:
type api struct { *App }


func (m *api) setApp(a *App) *api { m.App = a; return m }

func (m *api) getMuxRouter() *mux.Router {
	var r = mux.NewRouter()

	r.Host(m.conf.Base.Http.Host)
	r.Headers("Content-Type", "application/json")

	r.Use(m.httpMiddlewareRequestLog)
	r.Use(m.httpMiddlewareRequestCreate)

	r.HandleFunc("/v1", m.httpHandlerRootV1).Methods("GET")
	r.HandleFunc("/v1/alerts/{id:[0-9]+}", m.httpHandlerAlertsGet).Methods("GET")
	r.HandleFunc("/v1/alerts/{id:[0-9]+}", m.httpHandlerAlertsCreate).Methods("POST")

	return r
}

func (m *api) httpMiddlewareRequestLog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var errs = new(apiErrors).setApiPointer(m)
		var htRequest = new(httpRequest).create(m)

		errs.setRequestId(htRequest.id).setRequestLink(r.RequestURI)

		if e := htRequest.save(r); e != nil {
			m.log.Error().Err(e).Msg("Could not save request in the database!")
			errs.newError(errInternalCommonError) }

		context.Set(r, "internal_errors", errs)
		context.Set(r, "internal_request", htRequest)

		h.ServeHTTP(w,r)

		htRequest.saveStatus(errs.requestStatus)
	})
}
func (m *api) httpMiddlewareRequestCreate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.log.Debug().Msg("httpMiddlewareRequestCreate")

		h.ServeHTTP(w,r)

		m.log.Debug().Msg("httpMiddlewareRequestCreate2")
	})
}

func (m *api) httpHandlerRootV1(w http.ResponseWriter, r *http.Request) {}
func (m *api) httpHandlerAlertsGet(w http.ResponseWriter, r *http.Request) {}
func (m *api) httpHandlerAlertsCreate(w http.ResponseWriter, r *http.Request) {
	// TODO : PHONE MUST BE AS URL PARAM
	//var mv = mux.Vars(r)
	var errs = context.Get(r, "internal_errors").(*apiErrors)

	var req *apiPostRequest
	reqBody,e := ioutil.ReadAll(r.Body); if !m.errorHandler(w,e,errs) { return }
	e = json.Unmarshal(reqBody, req); if !m.errorHandler(w,e,errs) { return }

	context.Set(r, "internal_api", m)
	context.Set(r, "param_phone", req.Data.Attributes.Phone)
	context.Set(r, "param_alert", req.Data.Attributes.Alert)

	var rspPayload *apiResponse
	if usr := getUserByPhone(r); usr == nil {
		rspPayload.Data = usr.putAlertInQueue() }

	var status int
	rspPayload.Errors,status = errs.logAndSave().getErrorResponse()
	m.respondJSON(w, rspPayload, status)
}

func (m *api) errorHandler(w http.ResponseWriter, e error, errs *apiErrors) bool {
	if e == nil { return true }

	errs.newError(errInternalCommonError)
	m.log.Error().Err(e).Msg("Abnormal function result!")

	rsp,status := errs.logAndSave().getErrorResponse()
	m.respondJSON(w, &apiResponse{ Errors: rsp }, status)
	return false
}

func (m *api) respondJSON(w http.ResponseWriter, payload *apiResponse, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(payload)
}
