package app

import "encoding/hex"
import "crypto/sha256"
import "crypto/hmac"
import "bytes"
import "strings"
import "io/ioutil"
import "net/http"
import "encoding/json"
import "github.com/gorilla/mux"
import "github.com/gorilla/context"

import "github.com/satori/go.uuid"

// JSON response structs:
// Recomendations are taken from jsonapi.org:
type apiResponse struct {
	Data *responseData	`json:"data,omitempty"`
	Errors []*responseError	`json:"errors,omitempty"`
}
type responseData struct {
	Type string	`json:"type,omitempty"`
	Id string	`json:"id,omitempty"`
	Attributes *dataAttributes	`json:"attributes,omitempty"`
	Links *dataLinks	`json:"links,omitempty"`
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
	Code int										`json:"code,omitempty"`
	Status int									`json:"status,omitempty"`
	Title string								`json:"title,omitempty"`
	Detail string								`json:"detail,omitempty"`
	Source *errorSource					`json:"source,omitempty"`
	Links *dataLinks						`json:"links,omitempty"`
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

	stmt,e := m.ap.sqldb.Prepare("INSERT INTO requests (id,srcip,method,size,url,status,user_agent) VALUES (?,?,?,?,?,?,?)"); if e != nil { return e }
	defer stmt.Close()

	if _,e = stmt.Exec(m.id, strings.Split(r.RemoteAddr, ":")[0], r.Method, r.ContentLength, r.RequestURI, 0, r.UserAgent()); e != nil { return e }
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
	r.Headers("Content-Type", "application/json",
						"X-Requested-With", "XMLHttpRequest")
	r.Use(m.httpMiddlewareRequestLog)

	s := r.PathPrefix("/v1").Subrouter()
	s.Use(m.httpMiddlewareAPIAuthentication)
	s.HandleFunc("/alerts", m.httpHandlerAlertsCreate).Methods("POST")
	s.HandleFunc("/alerts/{id:[0-9]+}", m.httpHandlerAlertsGet).Methods("GET")

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

func (m *api) httpMiddlewareAPIAuthentication(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var errs = context.Get(r, "internal_errors").(*apiErrors)

		var bodyBuf bytes.Buffer
		bufSize,e := bodyBuf.ReadFrom(r.Body); if !m.errorHandler(w,e,errs) { return }
		r.Body.Close()

		mac := hmac.New(sha256.New, []byte(m.conf.Base.Api.Sign_Secret))
		macSize,e := mac.Write(bodyBuf.Bytes()); if !m.errorHandler(w,e,errs) { return }

		if r.ContentLength != bufSize || r.ContentLength != int64(macSize) {
			m.log.Warn().Msg("Different sizes in request, buffer and mac!") }

		expectedMAC := mac.Sum(nil)
		receivedMAC,e := hex.DecodeString(strings.Split(
			r.Header.Get("Authorization"), " ")[1]); if !m.errorHandler(w,e,errs) { return }

		m.log.Debug().Str("mac_expected", hex.EncodeToString(expectedMAC)).Str("mac_recevied", strings.Split(r.Header.Get("Authorization"), " ")[1]).Msg("HMAC signs comparison")

		if ! hmac.Equal(expectedMAC, receivedMAC) {
			errs.newError(errAlertsNotAuthorized)

			var status int
			var rspPayload = new(apiResponse)
			rspPayload.Errors,status = errs.logAndSave().getErrorResponse()
			m.respondJSON(w, rspPayload, status); return }

		r.Body = ioutil.NopCloser(bytes.NewReader(bodyBuf.Bytes()))
		h.ServeHTTP(w,r)
	})
}


func (m *api) httpHandlerRootV1(w http.ResponseWriter, r *http.Request) {}
func (m *api) httpHandlerAlertsGet(w http.ResponseWriter, r *http.Request) {}
func (m *api) httpHandlerAlertsCreate(w http.ResponseWriter, r *http.Request) {
	var errs = context.Get(r, "internal_errors").(*apiErrors)

	var req *apiPostRequest
	reqBody,e := ioutil.ReadAll(r.Body); if !m.errorHandler(w,e,errs) { return }
	e = json.Unmarshal(reqBody, &req); if !m.errorHandler(w,e,errs) { return }

	context.Set(r, "internal_api", m)
	context.Set(r, "param_phone", req.Data.Attributes.Phone)
	context.Set(r, "param_alert", req.Data.Attributes.Alert)

	var status int
	var rspPayload = new(apiResponse)
	if phone,err := parseRawPhone(req.Data.Attributes.Phone); err == errNotError {
		if usr := new(userModel).construct(r); usr.findUserByPhone(phone) {
			rspPayload.Data,status = usr.sendAlertWithResponse(req.Data.Attributes.Alert)
		}
	} else { errs.newError(err) }

	var errStatus int
	if rspPayload.Errors,errStatus = errs.logAndSave().getErrorResponse(); errStatus != 0 { status = errStatus }
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
