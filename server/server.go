package server

import (
	"golang.org/x/sys/unix"
)

func initializeSocket(port int) int {

	fd, err := bindMainFd(port)

	if err != nil {
		panic(err)
	}

	unix.Listen(fd, unix.SOMAXCONN)

	return fd
}

func ExecuteServer(port int) {

	fd := initializeSocket(port)

	eventLoopInit(fd)

}
