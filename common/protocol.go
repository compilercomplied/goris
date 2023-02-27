package common

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const PROTOCOL_HEADER int = 4
const MESSAGE_MAX_SIZE int = 4096
const MESSAGE_LENGTH int = PROTOCOL_HEADER + MESSAGE_MAX_SIZE + 1

func ReadFromBuffer(buffer *bytes.Buffer) (string, error) {

	header := make([]byte, PROTOCOL_HEADER)
	_, err := buffer.Read(header)
	if err != nil {
		return "", err
	}

	contentLength := binary.LittleEndian.Uint32(header)
	data := make([]byte, contentLength)
	_, err = buffer.Read(data)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func WriteToBuffer(msg string) (*bytes.Buffer, error) {
	msglength := len(msg)
	if msglength > MESSAGE_MAX_SIZE {
		return nil, errors.New("message too long")
	}

	header := make([]byte, PROTOCOL_HEADER)
	var headerv = uint32(len(msg))
	binary.LittleEndian.PutUint32(header, headerv)

	buffer := bytes.NewBuffer(header)
	_, err := buffer.Write([]byte(msg))

	if err != nil {
		return nil, err
	}

	return buffer, nil

}
