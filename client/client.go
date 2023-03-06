package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"goris/common"

	"golang.org/x/sys/unix"
)

func sendRequest(fd int, request *common.ProtocolRequest) {

	wbuffer, err := common.AppendToBuffer(request, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Request: '%s'\n", request.ToString())
	_, err = unix.Write(fd, wbuffer.Bytes())
	if err != nil {
		panic(err)
	}

	bufferlength := common.MESSAGE_MAX_SIZE + common.PROTOCOL_HEADER + 1
	buffer := bytes.NewBuffer(make([]byte, bufferlength))
	_, err = unix.Read(fd, buffer.Bytes())
	if err != nil {
		panic(err)
	}

	response, err := common.ReadFromBuffer(buffer)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response: '%s'\n", response.ToString())

}

func ExecuteClient(request *common.ProtocolRequest) {

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

	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to socket")

	sendRequest(fd, request)

}
