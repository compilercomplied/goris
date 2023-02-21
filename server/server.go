package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

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

	const PROTOCOL_HEADER int = 4
	const MAX_SIZE int = 4096

	for {
		connfd, _, err := unix.Accept(fd)

		if err != nil {
			panic(err)
		}
		fmt.Println("Connection made")

		// --- refactor block later --------------------------------------------
		buffer := make([]byte, MAX_SIZE+PROTOCOL_HEADER+1)
		nread, err := unix.Read(connfd, buffer)
		if nread < 0 {
			continue
		} else if err != nil {
			panic(err)
		}
		buf := bytes.NewBuffer(buffer)

		header := make([]byte, 4)
		_, err = buf.Read(header)
		if err != nil {
			panic(err)
		}

		contentLength := binary.LittleEndian.Uint32(header)
		data := make([]byte, contentLength)
		_, err = buf.Read(data)
		if err != nil {
			panic(err)
		}

		decodedPayload := string(data)
		fmt.Printf("msg from client: %s\n", decodedPayload)

		serverResponse := fmt.Sprintf("right back at you -> %s", decodedPayload)

		// codify response
		respheader := make([]byte, 4)
		var headerv uint32 = uint32(len(serverResponse))
		binary.LittleEndian.PutUint32(respheader, headerv)

		respbuffer := bytes.NewBuffer(respheader)
		if err != nil {
			panic(err)
		}
		_, err = respbuffer.Write([]byte(serverResponse))
		if err != nil {
			panic(err)
		}

		nwrite, err := unix.Write(connfd, respbuffer.Bytes())
		if err != nil {
			panic(err)
		} else if nwrite < 0 {
			panic(errors.New("error writing data"))
		}

		unix.Close(connfd)

		// ---------------------------------------------------------------------

	}

}

func ExecuteServer() {
	const PORT int = 5000
	fd, err := bindNewSocket(PORT)

	if err != nil {
		panic(err)
	}

	unix.Listen(fd, unix.SOMAXCONN)

	socketLoop(fd)

}
