package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"goris/common"

	"golang.org/x/sys/unix"
)

const POLLING_TIMEOUT_MS int = 3000

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

func setNonBlockingFd(fd int) error {
	ufd := uintptr(fd)

	flags, err := unix.FcntlInt(ufd, unix.F_GETFL, 0)

	if err != nil {
		panic(err)
	}

	flags |= unix.O_NONBLOCK

	_, err = unix.FcntlInt(ufd, unix.F_SETFL, flags)

	if err != nil {
		panic(err)
	}

	return nil
}

func tryOneRequest(connection *Connection) bool {

	msg, err := common.ReadFromBuffer(&connection.rbuf)

	if err != nil {
		return false
	}

	fmt.Println("Request received from client -> " + msg)

	wbuf, err := common.WriteToBuffer("echo: " + msg)
	if err != nil {
		fmt.Println("Error writing response to buffer" + msg)
		// return true anyway because we've finished reading the request
		return true
	}

	connection.state = RESPONSE_ST
	connection.wbuf = *wbuf

	return true
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

		rv, errno = unix.Write(connection.fd, connection.wbuf.Bytes())

		// Internally, Go's `read()` `executes Syscall()` which returns
		// `syscall.Errno` as the error value.
		breakloop = rv >= 0 && errno != unix.EINTR
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

		rv, errno = unix.Read(connection.fd, connection.rbuf.Bytes())

		// Internally, Go's `read()` `executes Syscall()` which returns
		// `syscall.Errno` as the error value.
		breakloop = rv >= 0 && errno != unix.EINTR
	}

	if rv < 0 && errno == unix.EAGAIN {
		return true, errors.New("EAGAIN")
	}

	if rv < 0 {
		connection.state = END_ST
		return true, errors.New("EAGAIN")
	}

	if rv == 0 {
		if connection.rbuf.Len() > 0 {
			fmt.Println("Unexpected EOF")
		} else {
			fmt.Println("EOF")
		}

		connection.state = END_ST
		return true, nil
	}

	if connection.state != REQUEST_ST {
		return true, errors.New("expected connection state request")
	}

	breakloop = false
	for loop := true; loop; loop = !breakloop {
		breakloop = tryOneRequest(connection)
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

func socketLoop(fd int) {

	err := setNonBlockingFd(fd)
	if err != nil {
		panic(err)
	}
	fmt.Println("nonblocking master fd")

	// Map that uses fd as key.
	connections := make(map[int32]*Connection)

	for {

		// Clear and initialize the polling data.
		pollArgs := make([]unix.PollFd, 0)

		// Anchor the master file descriptor to the first position.
		// We'll use this one to orchestrate the other fds.
		pfd := new(unix.PollFd)
		pfd.Fd = int32(fd)
		pfd.Events = unix.POLLIN
		pollArgs = append(pollArgs, *pfd)
		fmt.Println("init for loop done")

		for _, connection := range connections {
			if connection == nil {
				continue
			}

			/*************************************************************************
			PollFD struct looks like:
			--------------------------------------------------------------------------
			type PollFd struct {
				Fd      int32
				Events  int16
				Revents int16
			}
			--------------------------------------------------------------------------

			`Events` is an input parameter (we'll write it ourselves) whereas
			`Revents` is an output param, filled by the kernel. We'll read it when
			processing our connections.

			Read poll(2) man:
			https://man7.org/linux/man-pages/man2/poll.2.html
			*************************************************************************/

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

}

func ExecuteServer() {
	fd, err := bindNewSocket(common.DEF_SERVER_PORT)

	if err != nil {
		panic(err)
	}

	unix.Listen(fd, unix.SOMAXCONN)

	socketLoop(fd)

}
