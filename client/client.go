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
	wbuffer, err := common.WriteToBuffer(msg)
	if err != nil {
		panic(err)
	}

	_, err = unix.Write(fd, *wbuffer)
	if err != nil {
		panic(err)
	}

	buffer := common.InitializeReadBuffer()
	_, err = unix.Read(fd, buffer)
	if err != nil {
		panic(err)
	}

	response, err := common.ReadFromBuffer(&buffer)
	if err != nil {
		panic(err)
	}

	fmt.Printf("msg from server: %s\n", response)

}
