package common

import (
	"bytes"
	"encoding/binary"
)

const PROTOCOL_HEADER int = 4
const MESSAGE_MAX_SIZE int = 4096

func ReadFromBuffer(buffer *[]byte) (string, error) {
	buf := bytes.NewBuffer(*buffer)

	header := make([]byte, PROTOCOL_HEADER)
	_, err := buf.Read(header)
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

func WriteToBuffer(msg string) (*[]byte, error) {

	header := make([]byte, PROTOCOL_HEADER)
	var headerv = uint32(len(msg))
	binary.LittleEndian.PutUint32(header, headerv)

	buffer := bytes.NewBuffer(header)
	_, err := buffer.Write([]byte(msg))

	if err != nil {
		return nil, err
	}

	bytes := buffer.Bytes()

	return &bytes, nil

}

func InitializeReadBuffer() ([]byte) {
		return make([]byte, MESSAGE_MAX_SIZE+PROTOCOL_HEADER+1)
}
