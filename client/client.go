package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"goris/common"
	"math/rand"
	"sync"

	"golang.org/x/sys/unix"
)

var messages []string = []string{
	"Excuse me sir",
	"do you have a moment",
	"to talk about our lord and saviour",
	"Richard Stallman?",
}

func sendRequest(fd int) {

	msgIdx := rand.Intn(len(messages) - 1)

	msg := messages[msgIdx]

	wbuffer, err := common.AppendToBuffer(msg, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Request: '%s'\n", msg)
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

	fmt.Printf("Response: '%s'\n", response)

}

func ExecuteClient(requests int, port int) {

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
		Port: port,
		Addr: [4]byte(hostAddr),
	}

	err = unix.Connect(fd, &addr)

	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to socket")

	var waitGroup sync.WaitGroup
	for i := 0; i < requests; i++ {
		waitGroup.Add(1)

		go func(filedescriptor int) {
			defer waitGroup.Done()
			sendRequest(filedescriptor)

		}(fd)

	}

	waitGroup.Wait()
	unix.Close(fd)

}
