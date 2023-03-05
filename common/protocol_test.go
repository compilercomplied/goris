package common

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// --- READ --------------------------------------------------------------------
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

func Test_ReadMessageFromEmptyBuffer_Errors(t *testing.T) {

	// --- Setup -----------------------------------------------------------------
	buffer := bytes.NewBuffer(make([]byte, 10))

	// --- Execute ---------------------------------------------------------------
	_, err := ReadFromBuffer(buffer)

	// --- Assert ----------------------------------------------------------------
	if err == nil {
		t.Fatal("expected error but got none")
	}

	if err.Error() != E_NOREQUESTS {
		t.Fatalf("expected message '%s' but got '%s' instead", "no more messages", err.Error())
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

	// --- Assert ----------------------------------------------------------------
	nomsg, err := ReadFromBuffer(buffer)

	if nomsg == msg {
		t.Fatalf("expected no message but got '%s' instead", nomsg)

	}

	if err.Error() != "EOF" {
		t.Fatalf("expected 'EOF' as error but got '%'", err)
	}

}

// --- APPEND ------------------------------------------------------------------
func Test_AppendToNilBuffer_OK(t *testing.T) {
	const msg string = "hello"

	// --- Execute ---------------------------------------------------------------
	buffer, err := AppendToBuffer(msg, nil)

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

func Test_AppendToExistingBuffer_OK(t *testing.T) {
	const msg string = "hello"
	existingBuffer := bytes.NewBuffer(make([]byte, 0))
	existingBuffer.WriteByte(1)

	// --- Execute ---------------------------------------------------------------
	buffer, err := AppendToBuffer(msg, existingBuffer)
	// Remove the leading byte. If the data is not being appended it
	_, _ = existingBuffer.ReadByte()

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

func Test_AppendToBuffer_SizeLimit_Errors(t *testing.T) {

	const msg string = "hello"
	var longmessage string
	for i := 0; i < MESSAGE_MAX_SIZE; i++ {
		longmessage = msg + longmessage
	}

	// --- Execute ---------------------------------------------------------------
	_, err := AppendToBuffer(longmessage, nil)

	// --- Assert ----------------------------------------------------------------
	if err == nil {
		t.Fatal("expected error but got none")
	}

	if err.Error() != E_MSGLENGTH {
		t.Fatalf("expected message '%s' but got '%s' instead", msg, err.Error())
	}

}

// --- SLICE -------------------------------------------------------------------
func Test_SliceRequestFromBuffer_OK(t *testing.T) {

	const firstmsg string = "first message"
	const secondmsg string = "second message"

	// --- Setup -----------------------------------------------------------------
	buffer, _ := AppendToBuffer(firstmsg, nil)
	buffer, _ = AppendToBuffer(secondmsg, buffer)

	// --- Execute ---------------------------------------------------------------
	first, remaining, err := ReadRequestFromBuffer(buffer)

	if err != nil {
		t.Fatal(err)
	}

	// --- Assert ----------------------------------------------------------------

	firstResponse, _ := ReadFromBuffer(bytes.NewBuffer(first))
	secondResponse, _ := ReadFromBuffer(buffer)

	if !remaining {
		t.Fatalf("there were more requests that went undetected")
	}

	if firstResponse != firstmsg {
		t.Fatalf("expected message '%s' but got '%s' instead", firstmsg, firstResponse)
	}

	if secondResponse != secondmsg {
		t.Fatalf("expected message '%s' but got '%s' instead", secondmsg, secondResponse)
	}

}
