package server

import (
	"bytes"
	"errors"
)

// Dumb thing you have to do because of lack of GADTs or enums.
type ConnectionState string

const (
	REQUEST_ST  ConnectionState = "request"
	RESPONSE_ST ConnectionState = "response"
	END_ST      ConnectionState = "end"
)

type Connection struct {
	fd    int
	state ConnectionState
	rbuf  bytes.Buffer
	wbuf  bytes.Buffer
}

func NewConnection(fd int, state ConnectionState) (*Connection, error) {

	if fd < 0 {
		return nil, errors.New("invalid fd value")
	}

	connection := new(Connection)

	connection.fd = fd
	connection.state = state

	connection.rbuf = *bytes.NewBuffer(make([]byte, 0))
	connection.wbuf = *bytes.NewBuffer(make([]byte, 0))

	return connection, nil

}
