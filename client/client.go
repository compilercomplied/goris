package client

import (
	"encoding/binary"
	"fmt"
	"goris/common"

	"golang.org/x/sys/unix"
)

func ExecuteClient() {

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)

	if err != nil {
		panic(err)
	}
	if fd < 0 {
		fmt.Println("Bad file descriptor ", fd)
		return
	}

	hostAddr := make([]byte, 4)
	binary.BigEndian.PutUint32(hostAddr, 0)
	addr := unix.SockaddrInet4{
		Port: common.DEF_SERVER_PORT,
		Addr: [4]byte(hostAddr),
	}

	err = unix.Connect(fd, &addr)
	fmt.Println("Connected to socket")

	if err != nil {
		panic(err)
	}

	msg := "hello"
	err = common.WriteMessage(fd, msg)

	if err != nil {
		panic(err)
	}

	response, err := common.ReadMessage(fd)
	if err != nil {
		panic(err)
	}

	fmt.Printf("msg from server: %s", response)

}
