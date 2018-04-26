package app

import "errors"
import "net/http"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import "github.com/satori/go.uuid"


var (
	errApiCommonTypeInvalid = errors.New("The request type and the link are not the same!")
)
const (
	errInternalCommonError = uint8(iota)
	errInternalSqlError
	errAlertsCreatePhoneLength
	errAlertsCreatePhoneFormat
	errAlertsCreatePhoneNotFound
)
var apiErrorsTitle = map[uint8]string{
	errInternalCommonError: "Internal error",
	errInternalSqlError: "Internal database error",
	errAlertsCreatePhoneLength: "Abnormal phone parameter length",
	errAlertsCreatePhoneFormat: "Abnormal phone parameter format",
	errAlertsCreatePhoneNotFound: "Phone not found",
}
var apiErrorsDetail = map[uint8]string{
	errInternalCommonError: "The current request could not processed! Please, try again later!",
	errInternalSqlError: "The current request could not processed due to a database error. Please, try again later!",
	errAlertsCreatePhoneLength: "The phone must be between 10 and 12 characters in length!",
	errAlertsCreatePhoneFormat: "The phone should star with +79 or 89 digits!",
	errAlertsCreatePhoneNotFound: "The requested phone was not found in the database!",
}
var apiErrorsStatus = map[uint8]int{
	errInternalCommonError: http.StatusInternalServerError,
	errInternalSqlError: http.StatusInternalServerError,
	errAlertsCreatePhoneLength: http.StatusBadRequest,
	errAlertsCreatePhoneFormat: http.StatusBadRequest,
	errAlertsCreatePhoneNotFound: http.StatusNotFound,
}


type apiError struct {
	e uint8
	eId string
	srcParam string
}

func (m *apiError) setError(e uint8) *apiError { m.e = e; return m }
func (m *apiError) setParameter(p string) *apiError { m.srcParam = p; return m }
func (m *apiError) getId() string {
	if len(m.eId) == 0 { m.eId = uuid.NewV1().String() }
	return m.eId
}


type apiErrors struct {
	// internal pointers:
	ap *api

	// payload:
	requestId string
	requestLink string
	errors []*apiError
}

func (m *apiErrors) setApiPointer(a *api) *apiErrors { m.ap = a; return m }
func (m *apiErrors) setRequestLink(l string) *apiErrors { m.requestLink = l; return m }
func (m *apiErrors) setRequestId(i string) *apiErrors { m.requestId = i; return m }

func (m *apiErrors) newError(e uint8) *apiError {
	var err *apiError = new(apiError).setError(e)
	m.errors = append(m.errors, err)
	return err
}

func (m *apiErrors) getErrorResponse() ([]*responseError, int) {
	var respCode int
	var respErrors []*responseError

	for _,v := range m.errors {
		respErrors = append(respErrors, &responseError{
			Id: v.getId(),
			Code: int(v.e),
			Status: apiErrorsStatus[v.e],
			Title: apiErrorsTitle[v.e],
			Detail: apiErrorsDetail[v.e],
			Source: &errorSource{
				Parameter: v.srcParam },
			Links: &dataLinks{
				Self: m.requestLink } })

		if apiErrorsStatus[v.e] > respCode {
		 respCode = apiErrorsStatus[v.e] }
	}

	return respErrors, respCode
}

func (m *apiErrors) logAndSave() *apiErrors {
	m.ap.log.Debug().Msg("Saving error reports")

	stmt,e := m.ap.sqldb.Prepare("INSERT INTO errors VALUES (?,?,?,?,?,?,?,?)"); if e != nil {
		m.ap.log.Error().Err(e).Msg("Could not prepare DB statement!")
		m.newError(errInternalSqlError)
		return m }
	defer stmt.Close()

	for _,v := range m.errors {
		m.ap.log.Info().Str("request_link", m.requestLink).Int("http_code", apiErrorsStatus[v.e]).Str("error_id", v.getId()).Str("error_title", apiErrorsTitle[v.e]).Msg("")

		_,e = stmt.Exec(v.getId(), m.requestId, v.e, apiErrorsStatus[v.e], m.sqlNullString(v.srcParam), m.requestLink, apiErrorsTitle[v.e], apiErrorsDetail[v.e]); if e != nil {
			m.ap.log.Error().Err(e).Str("error_id", v.getId()).Msg("Could not write error report!") }
	}

	return m
}

func (m *apiErrors) sqlNullString(s string) *sql.NullString {
	if len(s) == 0 { return &sql.NullString{} }
	return &sql.NullString{
		String: s,
		Valid: true }
}
