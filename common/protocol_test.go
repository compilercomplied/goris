package common

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func Test_ReadFromBuffer_OK(t *testing.T) {

	const msg string = "hello"

	// --- Setup -----------------------------------------------------------------
	header := make([]byte, PROTOCOL_HEADER)
	var headerv = uint32(len(msg))
	binary.LittleEndian.PutUint32(header, headerv)

	buffer := bytes.NewBuffer(header)
	_, _ = buffer.Write([]byte(msg))

	buf := buffer.Bytes()

	// --- Execute ---------------------------------------------------------------
	parsedMessage, err := ReadFromBuffer(&buf)

	// --- Assert ----------------------------------------------------------------
	if err != nil {
		t.Fatal(err)
	}
	if parsedMessage != msg {
		t.Fatalf("Expected message '%s' but got '%s' instead", msg, parsedMessage)

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

	buf := bytes.NewBuffer(*buffer)

	header := make([]byte, PROTOCOL_HEADER)
	_, err = buf.Read(header)
	if err != nil {
		t.Fatal(err)
	}

	contentLength := binary.LittleEndian.Uint32(header)
	data := make([]byte, contentLength)
	_, err = buf.Read(data)
	if err != nil {
		t.Fatal(err)
	}

	parsedMessage := string(data)

	if parsedMessage != msg {
		t.Fatalf("Expected message '%s' but got '%s' instead", msg, parsedMessage)
	}

}
