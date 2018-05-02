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

// JSON response structs:
// Recomendations are taken from jsonapi.org:
type (
	apiResponse struct {
		Data *responseData          `json:"data,omitempty"`
		Errors []*responseError     `json:"errors,omitempty"`
	}
	responseData struct {
		Type string                 `json:"type,omitempty"`
		Id string                   `json:"id,omitempty"`
		Attributes *dataAttributes  `json:"attributes,omitempty"`
		Links *dataLinks            `json:"links,omitempty"`
	}
	dataAttributes struct {
		Phone string                `json:"phone,omitempty"`
		Alert string                `json:"alert,omitempty"`
		Status string               `json:"status,omitempty"`
	}
	dataLinks struct {
		Self string                 `json:"self,omitempty"`
	}
	responseError struct {
		Id string                   `json:"id,omitempty"`
		Code int                    `json:"code,omitempty"`
		Status int                  `json:"status,omitempty"`
		Title string                `json:"title,omitempty"`
		Detail string               `json:"detail,omitempty"`
		Source *errorSource         `json:"source,omitempty"`
		Links *dataLinks            `json:"links,omitempty"`
	}
	errorSource struct {
		Parameter string            `json:"parameter,omitempty"`
	}

	// JSON request structs:
	apiPostRequest struct {
		Data *requestData           `json:"data"`
	}
	requestData struct {
		Type string                 `json:"type"`
		Attributes *dataAttributes  `json:"attributes"`
	}
)

type api struct {}

func (m *api) getMuxRouter() *mux.Router {
	var r = mux.NewRouter()

	r.Host(globConfig.Base.Http.Host)
	r.Use(m.httpMiddlewareRequestLog)

	s := r.PathPrefix("/v1").Headers("Content-Type", "application/vnd.api+json").Subrouter()
	s.Use(m.httpMiddlewareAPIAuthentication)
	s.HandleFunc("/alerts", m.httpHandlerAlertsCreate).Methods("POST")
	s.HandleFunc("/alerts/{id:[0-9]+}", m.httpHandlerAlertsGet).Methods("GET")

	return r
}

func (m *api) httpMiddlewareRequestLog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		req,e := new(httpRequest).createAndSave(r); if e != nil {
			globLogger.Error().Err(e).Msg("[API]: Could not save request in the database!")
			req.newError(errInternalCommonError) }
		context.Set(r, "internal_request", req)

		h.ServeHTTP(w,r)

		req.updateAndSave()
	})
}

func (m *api) httpMiddlewareAPIAuthentication(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var req = context.Get(r, "internal_request").(*httpRequest)

		var bodyBuf bytes.Buffer
		bufSize,e := bodyBuf.ReadFrom(r.Body); if !m.errorHandler(w,e,req) { return }
		r.Body.Close()

		mac := hmac.New(sha256.New, []byte(globConfig.Base.Api.Sign_Secret))
		macSize,e := mac.Write(bodyBuf.Bytes()); if !m.errorHandler(w,e,req) { return }

		if r.ContentLength != bufSize || r.ContentLength != int64(macSize) {
			globLogger.Warn().Msg("[API]: Different sizes in request, buffer and mac!") }

		expectedMAC := mac.Sum(nil)
		receivedMAC,e := hex.DecodeString(strings.Split(
			r.Header.Get("Authorization"), " ")[1]); if !m.errorHandler(w,e,req) { return }

		globLogger.Debug().Str("mac_expected", hex.EncodeToString(expectedMAC)).Msg("[API]: HMAC sings comparison")
		globLogger.Debug().Str("mac_received", strings.Split(r.Header.Get("Authorization"), " ")[1]).Msg("[API]: HMAC signs comparison")

		if ! hmac.Equal(expectedMAC, receivedMAC) {
			req.newError(errAlertsNotAuthorized)
			m.respondJSON(w, req, nil, 0); return }

		r.Body = ioutil.NopCloser(bytes.NewReader(bodyBuf.Bytes()))
		h.ServeHTTP(w,r)
	})
}


func (m *api) httpHandlerRootV1(w http.ResponseWriter, r *http.Request) {}
func (m *api) httpHandlerAlertsGet(w http.ResponseWriter, r *http.Request) {}

func (m *api) httpHandlerAlertsCreate(w http.ResponseWriter, r *http.Request) {
	var req = context.Get(r, "internal_request").(*httpRequest)

	var postRequest *apiPostRequest
	reqBody,e := ioutil.ReadAll(r.Body); if !m.errorHandler(w,e,req) { return }
	e = json.Unmarshal(reqBody, &postRequest); if !m.errorHandler(w,e,req) { return }

	switch {
	case postRequest.Data == nil,
		postRequest.Data.Type == "",
		postRequest.Data.Attributes == nil,
		postRequest.Data.Attributes.Alert == "",
		postRequest.Data.Attributes.Phone == "":
			req.newError(errAlertsUnknownApiFormat)
			m.respondJSON(w, req, nil, 0); return
	default:
		globLogger.Debug().Msg("[API]: data checker is OK!")
	}

	context.Set(r, "param_phone", postRequest.Data.Attributes.Phone)
	context.Set(r, "param_alert", postRequest.Data.Attributes.Alert)

	var status int
	var payloadData *responseData
	if phone,err := parseRawPhone(postRequest.Data.Attributes.Phone); err == errNotError {
		if usr := new(userModel).construct(r); usr.findUserByPhone(phone) {
			payloadData,status = usr.sendAlertWithResponse(postRequest.Data.Attributes.Alert)
		}
	} else { req.newError(err) }

	m.respondJSON(w, req, payloadData, status)
}

func (m *api) errorHandler(w http.ResponseWriter, e error, req *httpRequest) bool {
	if e == nil { return true }

	req.newError(errInternalCommonError)
	globLogger.Error().Err(e).Msg("[API]: Abnormal function result!")

	m.respondJSON(w, req, nil, 0)
	return false
}

func (m *api) respondJSON(w http.ResponseWriter, req *httpRequest, payloadData *responseData, status int) {
	//
	var rspPayload = &apiResponse{
		Data: payloadData }

	if rspPayload.Errors,status = req.saveErrors().respondApiErrors(); req.status > status {
		status = req.status
		rspPayload.Data = nil	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(rspPayload)
}
