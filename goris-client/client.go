package main

import (
	"encoding/binary"
	"fmt"

	"golang.org/x/sys/unix"
)

func main() {

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

	nwrite, err := unix.Write(fd, []byte(msg))
	fmt.Println("sent msg")

	if nwrite < 0 {
		fmt.Println("Error writing")
		return
	}
	if err != nil {
		panic(err)
	}

	payload := make([]byte, 64)
	nread, err := unix.Read(fd, payload)
	if err != nil {
		panic(err)
	}
	if nread < 0 {
		fmt.Println("Error reading data")
		return
	}

	decodedPayload := string(payload)
	fmt.Printf("msg from server: %s", decodedPayload)

}
