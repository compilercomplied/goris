package server

import (
	"goris/common"

	"golang.org/x/sys/unix"
)

const POLLING_TIMEOUT_MS int = 3000



func initializeSocket(port int) (int) {

	fd, err := bindMasterFd(port)

	if err != nil {
		panic(err)
	}

	unix.Listen(fd, unix.SOMAXCONN)

	return fd
}

func ExecuteServer() {

	fd := initializeSocket(common.DEF_SERVER_PORT)

	eventLoopInit(fd)

}
