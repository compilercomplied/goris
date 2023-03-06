package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"goris/common"
)

// --- Type --------------------------------------------------------------------
type ProtocolResponse struct {
	Message string
}

func (res *ProtocolResponse) TotalLength() int {
	return len(res.Message)
}

func (res *ProtocolResponse) ToString() string {
	return res.Message
}

// --- Read --------------------------------------------------------------------
func ReadResponseFromBuffer(buffer *bytes.Buffer) (*ProtocolResponse, error) {

	header := make([]byte, PROTOCOL_HEADER)
	_, err := buffer.Read(header)
	if err != nil {
		return nil, err
	}

	contentLength := binary.LittleEndian.Uint32(header)
	if contentLength <= 0 {
		return nil, errors.New(common.E_NOREQUESTS)
	}

	data := make([]byte, contentLength)
	_, err = buffer.Read(data)
	if err != nil {
		return nil, err
	}

	response := new(ProtocolResponse)
	response.Message = string(data)

	return response, nil
}

// --- Write -------------------------------------------------------------------
func AppendResponseToBuffer(res *ProtocolResponse, buffer *bytes.Buffer) (*bytes.Buffer, error) {

	if buffer == nil {
		buffer = bytes.NewBuffer(make([]byte, 0))
	}

	reqLength := res.TotalLength()
	if uint32(reqLength) > MESSAGE_MAX_SIZE {
		return nil, errors.New(common.E_MSGLENGTH)
	}

	header := make([]byte, PROTOCOL_HEADER)
	var headerv = reqLength
	binary.LittleEndian.PutUint32(header, uint32(headerv))

	_, err := buffer.Write(header)
	if err != nil {
		return nil, err
	}

	_, err = buffer.Write([]byte(res.Message))
	if err != nil {
		return nil, err
	}

	return buffer, nil

}
