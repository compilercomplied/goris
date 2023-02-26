package common

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"
)

func Test_ReadMessageFromBuffer_OK(t *testing.T) {

	const msg string = "hello"

	// --- Setup -----------------------------------------------------------------
	header := make([]byte, PROTOCOL_HEADER)
	var headerv = uint32(len(msg))
	binary.LittleEndian.PutUint32(header, headerv)

	buffer := bytes.NewBuffer(header)
	_, _ = buffer.Write([]byte(msg))

	// --- Execute ---------------------------------------------------------------
	parsedMessage, err := ReadFromBuffer(buffer)

	// --- Assert ----------------------------------------------------------------
	if err != nil {
		t.Fatal(err)
	}
	if parsedMessage != msg {
		t.Fatalf("expected message '%s' but got '%s' instead", msg, parsedMessage)

	}

}

func Test_ReadMessageFromBuffer_MovesCursor(t *testing.T) {

	const msg string = "hello"

	// --- Setup -----------------------------------------------------------------
	header := make([]byte, PROTOCOL_HEADER)
	var headerv = uint32(len(msg))
	binary.LittleEndian.PutUint32(header, headerv)

	buffer := bytes.NewBuffer(header)
	_, _ = buffer.Write([]byte(msg))

	// --- Execute ---------------------------------------------------------------
	_, _ = ReadFromBuffer(buffer)
	nomsg, err := ReadFromBuffer(buffer)

	// --- Assert ----------------------------------------------------------------

	if nomsg == msg {
		t.Fatalf("expected no message but got '%s' instead", nomsg)
	}

	if err == nil {
		t.Fatalf("expected an error but got none")
	}

	if err.Error() != "EOF" {
		t.Fatalf("expected 'EOF' as error but got '%'", err)
	}

}

func Test_WriteToBuffer_OK(t *testing.T) {
	const msg string = "hello"

	// --- Execute ---------------------------------------------------------------
	buffer, err := WriteToBuffer(msg)

	// --- Assert ----------------------------------------------------------------
	if err != nil {
		t.Fatal(err)
	}

	header := make([]byte, PROTOCOL_HEADER)
	_, err = buffer.Read(header)
	if err != nil {
		t.Fatalf("found error '%s'", err.Error())
	}

	contentLength := binary.LittleEndian.Uint32(header)
	data := make([]byte, contentLength)
	_, err = buffer.Read(data)
	if err != nil {
		t.Fatalf("found error '%s'", err.Error())
	}

	parsedMessage := string(data)

	if parsedMessage != msg {
		t.Fatalf("expected message '%s' but got '%s' instead", msg, parsedMessage)
	}

}

func Test_WriteToBuffer_SizeLimit_Errors(t *testing.T) {
	const overflowingLength = MESSAGE_MAX_SIZE + 1
	var sb strings.Builder
	for i := 0; i < overflowingLength; i++ {
		sb.WriteString("x")
	}

	longmessage := sb.String()

	// --- Execute ---------------------------------------------------------------
	_, err := WriteToBuffer(longmessage)

	// --- Assert ----------------------------------------------------------------
	if err == nil {
		t.Fatalf("expected an error but got none")
	}

	errmsg := err.Error()
	if errmsg != "message too long" {
		t.Fatalf("expected error 'message too long' but got '%s' instead", errmsg)
	}

}
