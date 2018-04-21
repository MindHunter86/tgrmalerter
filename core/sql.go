package core

import "time"
import "database/sql"

import _ "github.com/go-sql-driver/mysql"
import mysql "github.com/go-sql-driver/mysql"

import "github.com/mattes/migrate"
import mysql_migrate "github.com/mattes/migrate/database/mysql"
import _ "github.com/mattes/migrate/source/file"


type sqlDriver interface {
	getRawDBSession() *sql.DB

	construct() (sqlDriver, error)
	destruct() error
}

type mysqlDriver struct {
	conf *CoreConfig

	sqlSession *sql.DB
	sqlMigration *migrate.Migrate
}


func (m *mysqlDriver) setConfig(c *CoreConfig) *mysqlDriver { m.conf = c; return m }
func (m *mysqlDriver) getRawDBSession() *sql.DB { return m.sqlSession }
func (m *mysqlDriver) destruct() error { return m.sqlSession.Close() }

func (m *mysqlDriver) construct() (sqlDriver, error) {
	if sess,e := sql.Open("mysql", m.connConfigure().FormatDSN()); e == nil {
		defer sess.Close()

		if err := m.migrationsRun(sess); err != nil { return nil,err }
	} else { return nil,e }

	return m,m.connCreate()
}

func (m *mysqlDriver) connCreate() error {
	var e error
	if m.sqlSession,e = sql.Open("mysql", m.connConfigure().FormatDSN()); e != nil { return e }
	return m.sqlSession.Ping()
}

func (m *mysqlDriver) connConfigure() *mysql.Config {
	// https://github.com/go-sql-driver/mysql - mysql lib configuration

	location,e := time.LoadLocation("Europe/Moscow"); if e != nil { location = time.UTC }

	return &mysql.Config{
		Net: "tcp",
		Addr: m.conf.Base.Mysql.Host,
		User: m.conf.Base.Mysql.Username,
		Passwd: m.conf.Base.Mysql.Password,
		DBName: m.conf.Base.Mysql.Database,
		Collation: "utf8_general_ci",
		MaxAllowedPacket: 0,
		TLSConfig: "false",
		Loc: location,

		Timeout: 10 * time.Second,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second,

		AllowAllFiles: false,
		AllowCleartextPasswords: false,
		AllowNativePasswords: false,
		AllowOldPasswords: false,
		ClientFoundRows: false,
		ColumnsWithAlias: false,
		InterpolateParams: false,
		MultiStatements: true,
		ParseTime: true,
		Strict: m.conf.Base.Mysql.Sql_Debug }
}

func (m *mysqlDriver) migrationsRun(sess *sql.DB) error {
	driver,e := mysql_migrate.WithInstance(sess, &mysql_migrate.Config{}); if e != nil { return e }
	if m.sqlMigration,e = migrate.NewWithDatabaseInstance("file://" + m.conf.Base.Mysql.Migrations_Path, "mysql", driver); e != nil { return e }

	if e = m.sqlMigration.Up(); e != nil && e != migrate.ErrNoChange { return e }
	return nil
}
