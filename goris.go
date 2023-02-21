package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func socketLoop(fd int) {

	for {
		connfd, _, err := unix.Accept(fd)

		if err != nil {
			panic(err)
		}
		fmt.Println("Connection made")

		if connfd < 0 {
			fmt.Println("Error: connfd is less than zero")
			continue
		}

		// --- refactor block later --------------------------------------------
		payload := make([]byte, 64)
		nread, err := unix.Read(connfd, payload)
		if err != nil {
			panic(err)
		}
		if nread < 0 {
			fmt.Println("Error reading data")
			continue
		}

		decodedPayload := string(payload)
		fmt.Printf("msg from client: %s\n", decodedPayload)

		serverResponse := fmt.Sprintf("right back at you -> %s", decodedPayload)
		nwrite, err := unix.Write(connfd, []byte(serverResponse))
		if err != nil {
			panic(err)
		}
		if nwrite < 0 {
			fmt.Println("Error writing data")
			continue
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
