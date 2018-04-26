package sql

import "database/sql"
import "mailru/tgrmalerter/core/config"


type SqlDriver interface {
	GetRawDBSession() *sql.DB

	Construct() (SqlDriver, error)
	Destruct() error

	SetConfig(*config.CoreConfig) SqlDriver
}
