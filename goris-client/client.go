package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"golang.org/x/sys/unix"
)

func main() {
	const PROTOCOL_HEADER int = 4
	const MAX_SIZE int = 4096

	const PORT int = 5000

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
		Port: PORT,
		Addr: [4]byte(hostAddr),
	}

	err = unix.Connect(fd, &addr)
	fmt.Println("Connected to socket")

	if err != nil {
		panic(err)
	}

	msg := "hello"

	header := make([]byte, 4)
	var headerv uint32 = uint32(len(msg))
	binary.LittleEndian.PutUint32(header, headerv)

	buffer := bytes.NewBuffer(header)
	if err != nil {
		panic(err)
	}
	_, err = buffer.Write([]byte(msg))
	if err != nil {
		panic(err)
	}

	_, err = unix.Write(fd, buffer.Bytes())
	fmt.Println("sent msg")
	if err != nil {
		panic(err)
	}

	respbuffer := make([]byte, MAX_SIZE+PROTOCOL_HEADER+1)
	nread, err := unix.Read(fd, respbuffer)
	if nread < 0 {
		panic(errors.New("nread is less than zero"))
	} else if err != nil {
		panic(err)
	}
	buf := bytes.NewBuffer(respbuffer)

	respheader := make([]byte, 4)
	_, err = buf.Read(respheader)
	if err != nil {
		panic(err)
	}

	contentLength := binary.LittleEndian.Uint32(respheader)
	data := make([]byte, contentLength)
	_, err = buf.Read(data)
	if err != nil {
		panic(err)
	}
	decodedPayload := string(data)
	fmt.Printf("msg from server: %s", decodedPayload)

}
