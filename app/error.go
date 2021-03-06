package app

import "errors"
import "net/http"
import "github.com/satori/go.uuid"

const (
	errNotError = uint8(iota)
	errInternalCommonError
	errInternalSqlError
	errAlertsNotAuthorized
	errAlertsUnknownApiFormat
	errAlertsUnknownType
	errAlertsCreatePhoneLength
	errAlertsCreatePhoneFormat
	errAlertsCreatePhoneNotFound
	errAlertsCreatePhoneNotRegistered
)
var (
	// common errors:
	errApiCommonTypeInvalid = errors.New("The request type and the link are not the same!")

	// api errors:
	apiErrorsTitle = map[uint8]string{
		errNotError: "",
		errInternalCommonError: "Internal error",
		errInternalSqlError: "Internal database error",
		errAlertsNotAuthorized: "Authorization failed",
		errAlertsUnknownApiFormat: "Unknown API request format",
		errAlertsUnknownType: "Unknown request type",
		errAlertsCreatePhoneLength: "Abnormal phone parameter length",
		errAlertsCreatePhoneFormat: "Abnormal phone parameter format",
		errAlertsCreatePhoneNotFound: "Phone not found",
		errAlertsCreatePhoneNotRegistered: "Inactive user found",
	}
	apiErrorsDetail = map[uint8]string{
		errNotError: "",
		errInternalCommonError: "The current request could not processed! Please, try again later!",
		errInternalSqlError: "The current request could not processed due to a database error. Please, try again later!",
		errAlertsNotAuthorized: "The current request must be signed with a special key for correct authorization! Please, check your credentials!",
		errAlertsUnknownApiFormat: "Could not parse request! Please read the documentation and try again!",
		errAlertsUnknownType: "The current request has a type that was sent incorrectly!",
		errAlertsCreatePhoneLength: "The phone must be between 10 and 12 characters in length!",
		errAlertsCreatePhoneFormat: "The phone should star with +79 or 89 digits!",
		errAlertsCreatePhoneNotFound: "The requested phone was not found in the database!",
		errAlertsCreatePhoneNotRegistered: "The requested phone was found but it refers to the disabled user!",
	}
	apiErrorsStatus = map[uint8]int{
		errNotError: 0,
		errInternalCommonError: http.StatusInternalServerError,
		errInternalSqlError: http.StatusInternalServerError,
		errAlertsNotAuthorized: http.StatusUnauthorized,
		errAlertsUnknownApiFormat: http.StatusBadRequest,
		errAlertsUnknownType: http.StatusBadRequest,
		errAlertsCreatePhoneLength: http.StatusBadRequest,
		errAlertsCreatePhoneFormat: http.StatusBadRequest,
		errAlertsCreatePhoneNotFound: http.StatusNotFound,
		errAlertsCreatePhoneNotRegistered: http.StatusLocked,
	}
)


type apiError struct {
	e uint8
	eId string
	srcParam string
}

func (m *apiError) setError(e uint8) *apiError { m.e = e; return m }
func (m *apiError) setParameter(p string) *apiError { m.srcParam = p; return m }
func (m *apiError) getId() string {
	if len(m.eId) == 0 { m.eId = uuid.NewV4().String() }
	return m.eId
}
