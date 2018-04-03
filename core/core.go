package core

import (
	"errors"
	"github.com/rs/zerolog"
)

type Core struct {
	log *zerolog.Logger
	cfg *CoreConfig
}

var (
	errPathLenZero = errors.New("Given config path argument has len 0!")
	errPathNotExist = errors.New("Given config path argument is invalid! No such file or directory!")
)


func (m *Core) SetLogger(l *zerolog.Logger) *Core { m.log = l; return m }
func (m *Core) SetConfig(c *CoreConfig) *Core { m.cfg = c; return m }
func (m *Core) Construct() (*Core, error) {
	return m,nil
}

func (m *Core) Bootstrap() error { return nil }
