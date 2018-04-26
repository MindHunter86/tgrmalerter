package app

import "io/ioutil"
import "net/http"
import "encoding/json"
import "github.com/gorilla/mux"
import "github.com/gorilla/context"

// JSON response structs:
// Recomendations are taken from jsonapi.org:
type apiResponse struct {
	Data *responseData					`json:data,omitempty`
	Errors []*responseError			`json:errors,omitempty`
}
type responseData struct {
	Type string									`json:type,omitempty`
	Id int											`json:id,omitempty`
	Attributes *dataAttributes	`json:attributes,omitempty`
	Links *dataLinks						`json:links,omitempty`
}
type dataAttributes struct {
	Phone string								`json:phone,omitempty`
	Alert string								`json:alert,omitempty`
	Status string								`json:status,omitempty`
}
type dataLinks struct {
	Self string									`json:self,omitempty`
}
type responseError struct {
	Id string										`json:id,omitempty`
	Links *dataLinks						`json:links,omitempty`
	Code int										`json:code,omitempty`
	Status int									`json:status,omitempty`
	Title string								`json:title,omitempty`
	Detail string								`json:detail,omitempty`
	Source *errorSource					`json:source,omitempty`
}
type errorSource struct {
	Parameter string						`json:parameter,omitempty`
}

// JSON request structs:
type apiPostRequest struct {
	Data *requestData						`json:data,omitempty`
}
type requestData struct {
	Type string									`json:type,omitempty`
	Attributes *dataAttributes	`json:attributes,omitempty`
}

// request struct:
// TODO:
//  - check type and link
type httpRequest struct {
	// internal pointers:
	ap *api

	// payload:
	id string
	link string
}

func (m *httpRequest) create() *httpRequest { return nil }

// internal Application struct:
type api struct { *App }
type alertsHandler struct { ap *api }


func (m *api) setApp(a *App) *api { m.App = a; return m }

func (m *api) getMuxRouter() *mux.Router {
	var r = mux.NewRouter()
	var htAlerts = &alertsHandler{m}

	r.Host(m.conf.Base.Http.Host)
	r.Headers("Content-Type", "application/json")

	// TODO: add status && version
//	r.HandleFunc("/", m.httpRootHandler).Methods("GET")
	r.HandleFunc("/v1/alerts/{id:[0-9]+}", htAlerts.Get).Methods("GET")
	r.HandleFunc("/v1/alerts/{id:[0-9]+}", htAlerts.Create).Methods("POST")

	return r
}

func (m *alertsHandler) Get(w http.ResponseWriter, r *http.Request) {}

func (m *alertsHandler) Create(w http.ResponseWriter, r *http.Request) {
	// TODO : PHONE MUST BE AS URL PARAM
	//var mv = mux.Vars(r)
	var errs = new(apiErrors)

	var req *apiPostRequest
	reqBody,e := ioutil.ReadAll(r.Body); if !m.ap.errorHandler(w,e,errs) { return }
	e = json.Unmarshal(reqBody, req); if !m.ap.errorHandler(w,e,errs) { return }

	context.Set(r, "inter_api", m.ap)
	context.Set(r, "inter_errs", errs)
	context.Set(r, "param_phone", req.Data.Attributes.Phone)
	context.Set(r, "param_alert", req.Data.Attributes.Alert)

	var rspPayload *apiResponse
	if usr := getUserByPhone(r); usr == nil {
		rspPayload.Data = usr.putAlertInQueue() }

	var status int
	rspPayload.Errors,status = errs.logAndSave().getErrorResponse()
	m.ap.respondJSON(w, rspPayload, status)
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
