package server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"goris/common"

	"golang.org/x/sys/unix"
)

func bindNewSocket(port int) (fd int, err error) {

	// AF_INET:
	//  - IPv4 Internet protocols.
	// SOCK_STREAM:
	//  - Used for TCP (whereas SOCK_DGRAM is used for UDP).
	// int 0:
	//  - https://stackoverflow.com/a/3735791
	fd, err = unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)

	if err != nil {
		fmt.Println("error while creating file descriptor")
		panic(err)
	}
	if fd < 0 {
		panic(errors.New("malformed file descriptor returned"))
	}

	// SOL_SOCKET:
	// SO_REUSEADDR:
	//  - https://pubs.opengroup.org/onlinepubs/7908799/xns/getsockopt.html
	val := 1
	err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, val)

	if err != nil {
		fmt.Println("error while setting sock opts")
		panic(err)
	}

	// Set host address as *0.0.0.0*. Code in C specifies using ntohl(3) syscall
	// to get host byte order from the uint. Network byte order is always
	// big endian.
	hostAddr := make([]byte, 4)
	binary.BigEndian.PutUint32(hostAddr, 0)
	addr := unix.SockaddrInet4{
		Port: port,
		Addr: [4]byte(hostAddr),
	}

	err = unix.Bind(fd, &addr)

	if err != nil {
		fmt.Println("error while binding fd to address")
		panic(err)
	}

	return fd, nil

}
func socketLoop(fd int) {
	const max_connections int = 1
	var current_connections int = 0
	for {
		if current_connections == max_connections {
			fmt.Println("Max connections for server reached")
			break
		}

		connfd, _, err := unix.Accept(fd)

		if err != nil {
			panic(err)
		}
		fmt.Println("Connection made")

		message, err := common.ReadMessage(connfd)
		if err != nil {
			panic(err)
		}

		fmt.Printf("msg from client: %s\n", message)

		serverResponse := fmt.Sprintf("right back at you -> %s", message)

		err = common.WriteMessage(connfd, serverResponse)
		if err != nil {
			panic(err)
		}

		unix.Close(connfd)
		current_connections++

	}

}

func ExecuteServer() {
	fd, err := bindNewSocket(common.DEF_SERVER_PORT)

	if err != nil {
		panic(err)
	}

	unix.Listen(fd, unix.SOMAXCONN)

	socketLoop(fd)

}
