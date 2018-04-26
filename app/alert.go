package app

import ""

type mdlAlert struct {
  log *zerolog.Logger

  msg, phone []byte
}

func (m *mdlAlert) construct(numer string) (*mdlAlert,error) {
	return m,nil
}
