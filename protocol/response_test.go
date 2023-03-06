package protocol

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func defaultResponse(msg string) (*ProtocolResponse, *bytes.Buffer) {
	resp := new(ProtocolResponse)
	resp.Message = msg

	buffer := bytes.NewBuffer(make([]byte, 0))

	respLength := resp.TotalLength()
	header := make([]byte, PROTOCOL_HEADER)
	var headerv = respLength
	binary.LittleEndian.PutUint32(header, uint32(headerv))

	buffer.Write(header)
	buffer.Write([]byte(resp.Message))

	return resp, buffer
}

func Test_ReadResponseFromBuffer_OK(t *testing.T) {
	// --- Arrange ---------------------------------------------------------------
	const msg = "hello"
	resp, buffer := defaultResponse(msg)
	// --- Act -------------------------------------------------------------------

	parsed, err := ReadResponseFromBuffer(buffer)
	// --- Assert ----------------------------------------------------------------
	if err != nil {
		t.Fatal(err)
	}

	if resp.Message != parsed.Message {
		t.Fatalf("expected message '%s' but got '%s' instead", resp.Message, parsed.Message)
	}

}

func Test_AppendResponseToNilBuffer_OK(t *testing.T) {
	// --- Arrange ---------------------------------------------------------------
	const msg = "hello"
	resp, originalBuffer := defaultResponse(msg)
	// --- Act -------------------------------------------------------------------

	buffer, err := AppendResponseToBuffer(resp, nil)
	// --- Assert ----------------------------------------------------------------
	if err != nil {
		t.Fatal(err)
	}

	if buffer.Bytes()[0] != originalBuffer.Bytes()[0] {
		t.Fatalf("response not codified correctly")
	}

}

func Test_AppendResponseToExistingBuffer_OK(t *testing.T) {
	// --- Arrange ---------------------------------------------------------------
	const msg = "hello"
	resp, originalBuffer := defaultResponse(msg)
	// --- Act -------------------------------------------------------------------

	buffer, err := AppendResponseToBuffer(resp, bytes.NewBuffer(make([]byte,0)))
	// --- Assert ----------------------------------------------------------------
	if err != nil {
		t.Fatal(err)
	}

	if buffer.Bytes()[0] != originalBuffer.Bytes()[0] {
		t.Fatalf("response not codified correctly")
	}

}
