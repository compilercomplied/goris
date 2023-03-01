package server

import (
	"bytes"
	"errors"
	"fmt"
	"goris/common"

	"golang.org/x/sys/unix"
)

func tryOneRequest(connection *Connection) bool {

	msg, err := common.ReadFromBuffer(&connection.rbuf)

	if err != nil {
		return true
	}

	drainedBuffer := bytes.NewBuffer(connection.rbuf.Bytes())
	connection.rbuf = *drainedBuffer

	fmt.Println("Request received from client -> " + msg)

	wbuf, err := common.AppendToBuffer("echo: "+msg, &connection.wbuf)
	if err != nil {
		fmt.Println("Error writing response to buffer" + msg)
		// Return true anyway because we've finished reading the request.
		return true
	}

	connection.state = RESPONSE_ST
	connection.wbuf = *wbuf

	return false
}

func sendResponse(connection *Connection) {

	breakloop := false
	for loop := true; loop; loop = !breakloop {
		breakloop = !tryFlushBuffer(connection)
	}

}

func tryFlushBuffer(connection *Connection) bool {

	breakloop := false
	var rv int
	var errno error
	for loop := true; loop; loop = !breakloop {

		response, remaining, err := common.ReadRequestFromBuffer(&connection.wbuf)

		if err != nil {
			break
		}

		rv, errno = unix.Write(connection.fd, response)

		// Internally, Go's `read()` `executes Syscall()` which returns
		// `syscall.Errno` as the error value.
		breakloop = rv >= 0 && errno != unix.EINTR && !remaining
	}

	if rv < 0 && errno == unix.EAGAIN {
		return false
	}

	if rv < 0 {
		fmt.Println("Error writing back response")
		connection.state = END_ST
		return false
	}

	connection.state = REQUEST_ST
	connection.wbuf = *bytes.NewBuffer(make([]byte, common.MESSAGE_MAX_SIZE))

	return false
}

func tryFillBuffer(connection *Connection) (bool, error) {

	// Read the whole fd buffer into our connection read buffer.
	breakloop := false
	var rv int
	var errno error
	for loop := true; loop; loop = !breakloop {
		buffer := make([]byte, 64)
		rv, errno = unix.Read(connection.fd, buffer)

		// Internally, Go's `read()` `executes Syscall()` which returns
		// `syscall.Errno` as the error value.
		breakloop = rv <= 0 && errno != unix.EINTR

		if rv > 0 {
			connection.rbuf.Write(buffer)
		}
	}

	if rv < 0 && errno == unix.EAGAIN {
		return true, nil
	}

	if rv < 0 && errno != nil {
		connection.state = END_ST
		return true, errno
	} else if rv < 0 {
		return true, errors.New("unexpected negative unix.Read() without errors")
	}

	if rv == 0 {
		// No more data to be read on this buffer.
		connection.state = END_ST
		return true, nil
	}

	return true, nil

}

func handleRequest(connection *Connection) error {

	breakloop := false
	var err error = nil
	for loop := true; loop; loop = !breakloop {
		finished, err := tryFillBuffer(connection)
		if err != nil {
			breakloop = true
		}
		if finished {
			break
		}
	}

	if err != nil {
		return err
	}

	breakloop = false
	for loop := true; loop; loop = !breakloop {
		breakloop = tryOneRequest(connection)
	}

	return err

}

func mutateConnectionBuffers(connection *Connection) error {
	switch connection.state {
	case REQUEST_ST:
		handleRequest(connection)
		return nil
	case RESPONSE_ST:
		sendResponse(connection)
		return nil
	default:
		msg := fmt.Sprintf("unknown request state '%v'", connection.state)
		return errors.New(msg)
	}
}
func acceptNewConnection(fd int) (*Connection, error) {

	connfd, _, err := unix.Accept(fd)

	if err != nil {
		return nil, err
	}

	err = setNonBlockingFd(fd)
	if err != nil {
		return nil, err
	}

	connection, err := NewConnection(connfd, REQUEST_ST)
	if err != nil {
		return nil, err
	}

	return connection, nil

}

func loopCycle(mainFd int, connections map[int32]*Connection) {

	// Initialize the polling data.
	pollArgs := make([]unix.PollFd, 0)

	// Anchor our main file descriptor to the first position.
	// We'll use this one to orchestrate the other fds.
	pfd := new(unix.PollFd)
	pfd.Fd = int32(mainFd)
	pfd.Events = unix.POLLIN
	pollArgs = append(pollArgs, *pfd)
	fmt.Println("init for loop done")

	for _, connection := range connections {
		if connection == nil {
			continue
		}

		/***************************************************************************
		PollFD struct looks like:
		----------------------------------------------------------------------------
		type PollFd struct {
			Fd      int32
			Events  int16
			Revents int16
		}
		----------------------------------------------------------------------------

		`Events` is an input parameter (we'll write it ourselves) whereas
		`Revents` is an output param, filled by the kernel. We'll read it when
		processing our connections.

		Read poll(2) man:
		https://man7.org/linux/man-pages/man2/poll.2.html
		***************************************************************************/

		pollfd := new(unix.PollFd)

		var pollState int16 = unix.POLLERR

		switch connection.state {
		case REQUEST_ST:
			pollState = unix.POLLIN
		case RESPONSE_ST:
			pollState = unix.POLLOUT
		default:
			pollState = unix.POLLERR
		}
		pollfd.Fd = int32(connection.fd)
		pollfd.Events = pollState

		pollArgs = append(pollArgs, *pollfd)
		fmt.Println("appended connection to pollArgs")

	}

	fmt.Printf("polling with pollargs length '%v'\n", len(pollArgs))
	_, err := unix.Poll(pollArgs, POLLING_TIMEOUT_MS)
	if err != nil {
		// TODO: handle some of its panics.
		// panic: interrupted system call
		panic(err)
	}

	fmt.Println("polled pollargs")

	// Be sure to skip the first one, since it is our main file desc.
	for i := 1; i < len(pollArgs); i++ {
		/*
			Both `Events` and `Revents` are bitfields. `poll.h` defines the
			following possible values:

			// These events are set in our `Events` field, but can also be polled
			// for (read from `Revents`).
			#define POLLIN   0x0001 // There is data to be read.
			#define POLLPRI  0x0002 // There is urgent data to be read.
			#define POLLOUT  0x0004 // Writing now will not block.

			// These events are always used for polling. They indicate the status
			// of the request.
			#define POLLERR  0x0008 // Error.
			#define POLLHUP  0x0010 // Hung up.
			#define POLLNVAL 0x0020 // Invalid polling request.
		*/

		currentPoll := pollArgs[i]
		if currentPoll.Revents > 0 {
			fmt.Println("Revents for connection")

			connection := connections[currentPoll.Fd]
			err = mutateConnectionBuffers(connection)
			if err != nil {
				panic(err)
			}

			if connection.state == END_ST {
				err = unix.Close(connection.fd)
				if err != nil {
					panic(err)
				}

				delete(connections, int32(connection.fd))
			}

		} else {
			fmt.Println("no Revents")
		}
	}

	masterPollFd := pollArgs[0]
	if masterPollFd.Revents > 0 {
		fmt.Println("Revents on master fd")
		connection, err := acceptNewConnection(int(masterPollFd.Fd))
		if err != nil {
			panic(err)
		}

		connections[int32(connection.fd)] = connection
	}

}

func eventLoopInit(fd int) {

	err := setNonBlockingFd(fd)
	if err != nil {
		panic(err)
	}
	fmt.Println("nonblocking master fd")

	// Map that uses fd as key.
	connections := make(map[int32]*Connection)

	for {
		loopCycle(fd, connections)
	}

}
