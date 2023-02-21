package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"golang.org/x/sys/unix"
)

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

func main() {
	const PORT int = 5000

	fd, err := bindNewSocket(PORT)

	if err != nil {
		panic(err)
	}

	unix.Listen(fd, unix.SOMAXCONN)

	socketLoop(fd)

}
