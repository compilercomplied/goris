package common

import (
	"bytes"
	"encoding/binary"

	"golang.org/x/sys/unix"
)

const PROTOCOL_HEADER int = 4
const MESSAGE_MAX_SIZE int = 4096

func ReadMessage(fd int) (string, error) {
	buffer := make([]byte, MESSAGE_MAX_SIZE+PROTOCOL_HEADER+1)
	_, err := unix.Read(fd, buffer)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(buffer)

	header := make([]byte, 4)
	_, err = buf.Read(header)
	if err != nil {
		return "", err
	}

	contentLength := binary.LittleEndian.Uint32(header)
	data := make([]byte, contentLength)
	_, err = buf.Read(data)
	if err != nil {
		return "", err
	}

	return string(data), nil

}

func WriteMessage(fd int, msg string) error {
	// TODO: validate size
	header := make([]byte, 4)
	var headerv uint32 = uint32(len(msg))
	binary.LittleEndian.PutUint32(header, headerv)

	buffer := bytes.NewBuffer(header)
	_, err := buffer.Write([]byte(msg))
	if err != nil {
		return err
	}

	_, err = unix.Write(fd, buffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}
