package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"goris/common"
)

// --- Type --------------------------------------------------------------------
type ProtocolRequest struct {
	Action string
	Key    string
	Value  *string
}

func (req *ProtocolRequest) TotalLength() uint32 {
	if req.Value != nil {
		return uint32(len(req.Action)+len(req.Key)+len(*(req.Value))) + (FRAGMENT_HEADER * 3) + PROTOCOL_HEADER
	} else {
		return uint32(len(req.Action)+len(req.Key)) + (FRAGMENT_HEADER * 2) + PROTOCOL_HEADER
	}
}

func (req *ProtocolRequest) ToString() string {
	if req.Value == nil {
		return fmt.Sprintf("[%v] => ['%v']", req.Action, req.Key)
	} else {
		return fmt.Sprintf("[%v] => ['%v']:['%v']", req.Action, req.Key, *(req.Value))
	}
}

func NewProtocolRequest(action string, key string, value *string) (*ProtocolRequest, error) {

	switch action {
	case "s":
		if key == "" {
			return nil, errors.New(common.E_INVALIDKEY)
		} else if value == nil || *value == "" {
			return nil, errors.New(common.E_INVALIDVALUE)
		} else {
			protocolReq := new(ProtocolRequest)
			protocolReq.Action = action
			protocolReq.Key = key
			protocolReq.Value = value
			return protocolReq, nil
		}
	case "g":
		if key == "" {
			return nil, errors.New(common.E_INVALIDKEY)
		} else {
			protocolReq := new(ProtocolRequest)
			protocolReq.Action = action
			protocolReq.Key = key
			protocolReq.Value = value
			return protocolReq, nil
		}
	case "d":
		if key == "" {
			return nil, errors.New(common.E_INVALIDKEY)
		} else {
			protocolReq := new(ProtocolRequest)
			protocolReq.Action = action
			protocolReq.Key = key
			protocolReq.Value = value
			return protocolReq, nil
		}
	default:
		return nil, fmt.Errorf(common.E_UNKNOWNREQ, action)
	}
}

// --- Read --------------------------------------------------------------------
func ReadRequestFromBuffer(buffer *bytes.Buffer) (data []byte, remaining bool, err error) {

	header := make([]byte, PROTOCOL_HEADER)
	_, err = buffer.Read(header)
	if err != nil {
		return data, remaining, err
	}

	contentLength := binary.LittleEndian.Uint32(header)
	if contentLength <= 0 {
		return data, remaining, errors.New(common.E_NOREQUESTS)
	}
	data = make([]byte, contentLength)

	_, err = buffer.Read(data)
	if err != nil {
		return data, remaining, err
	}

	rv, err := buffer.ReadByte()
	if rv == 0 || err != nil {
		remaining = false
	} else {
		remaining = true
		buffer.UnreadByte()
	}

	// Return the fully formed response; [header][data].
	return append(header, data...), remaining, nil
}

func ReadFromBuffer(buffer *bytes.Buffer) (*ProtocolRequest, error) {

	// Extract the whole request first. It has all the fragments inside.
	data, _, err := ReadRequestFromBuffer(buffer)
	if err != nil {
		return nil, err
	}

	// Skip the header; we only want the fragments inside the payload.
	requestBuffer := bytes.NewBuffer(data[PROTOCOL_HEADER:])

	fragmentsRaw := make([]byte, FRAGMENT_HEADER)
	_, err = requestBuffer.Read(fragmentsRaw)
	if err != nil {
		return nil, err
	}

	fragments := binary.LittleEndian.Uint16(fragmentsRaw)

	// Parse all the fragments before feeding them to our ProtocolRequest.
	// Maintain the same write order Action-Key-Value?.
	if fragments != 2 && fragments != 3 {
		return nil, errors.New(common.E_REQLENGTH)
	}

	action, err := readFragment(requestBuffer)
	if err != nil {
		return nil, err
	}
	key, err := readFragment(requestBuffer)
	if err != nil {
		return nil, err
	}
	var value *string
	if fragments == 3 {
		val, err := readFragment(requestBuffer)
		if err != nil {
			return nil, err
		}

		value = &val
	}

	return NewProtocolRequest(action, key, value)
}

func readFragment(buffer *bytes.Buffer) (string, error) {

	header := make([]byte, FRAGMENT_HEADER)
	_, err := buffer.Read(header)
	if err != nil {
		return "", err
	}

	contentLength := binary.LittleEndian.Uint32(header)
	if contentLength <= 0 {
		return "", err
	}

	data := make([]byte, contentLength)

	_, err = buffer.Read(data)
	if err != nil {
		return "", err
	}

	return string(data), err
}

// --- Write -------------------------------------------------------------------
func AppendToBuffer(req *ProtocolRequest, buffer *bytes.Buffer) (*bytes.Buffer, error) {

	if buffer == nil {
		buffer = bytes.NewBuffer(make([]byte, 0))
	}

	reqLength := req.TotalLength()
	if reqLength > MESSAGE_MAX_SIZE {
		return nil, errors.New(common.E_MSGLENGTH)
	}

	header := make([]byte, PROTOCOL_HEADER)
	var headerv = reqLength
	binary.LittleEndian.PutUint32(header, headerv)

	_, err := buffer.Write(header)
	if err != nil {
		return nil, err
	}

	fragmentNo := uint32(2)
	if req.Action == "s" {
		fragmentNo = 3
	}
	fragmentheader := make([]byte, FRAGMENT_HEADER)
	binary.LittleEndian.PutUint32(fragmentheader, fragmentNo)

	// Maintain the same read order Action-Key-Value?.
	_, err = buffer.Write(fragmentheader)
	if err != nil {
		return nil, err
	}

	err = writeFragment(req.Action, buffer)
	if err != nil {
		return nil, err
	}

	err = writeFragment(req.Key, buffer)
	if err != nil {
		return nil, err
	}

	if fragmentNo == 3 {
		err = writeFragment(*req.Value, buffer)
		if err != nil {
			return nil, err
		}
	}

	return buffer, nil
}

func writeFragment(msg string, buffer *bytes.Buffer) error {

	header := make([]byte, FRAGMENT_HEADER)
	var headerv = len(msg)
	binary.LittleEndian.PutUint32(header, uint32(headerv))

	_, err := buffer.Write(header)
	if err != nil {
		return err
	}

	_, err = buffer.Write([]byte(msg))
	if err != nil {
		return err
	}

	return nil
}
