package common

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func defaultReadSetup(action string, value *string) (ProtocolRequest, *bytes.Buffer) {

	var request *ProtocolRequest
	fragmentLengthv := uint32(3)
	if action == "s" {
		request, _ = NewProtocolRequest(action, "myKey", value)
		fragmentLengthv = uint32(3)
	} else {
		request, _ = NewProtocolRequest(action, "myKey", nil)
		fragmentLengthv = 2
	}

	header := make([]byte, PROTOCOL_HEADER)
	var headerv = request.TotalLength()
	binary.LittleEndian.PutUint32(header, headerv)
	buffer := bytes.NewBuffer(header)

	fragmentlength := make([]byte, FRAGMENT_HEADER)
	binary.LittleEndian.PutUint32(fragmentlength, fragmentLengthv)
	_, _ = buffer.Write(fragmentlength)

	actionLength := make([]byte, FRAGMENT_HEADER)
	var actionLengthv = uint32(len(request.Action))
	binary.LittleEndian.PutUint32(actionLength, actionLengthv)
	_, _ = buffer.Write(actionLength)
	_, _ = buffer.Write([]byte(request.Action))

	keylength := make([]byte, FRAGMENT_HEADER)
	var keylengthv = uint32(len(request.Key))
	binary.LittleEndian.PutUint32(keylength, keylengthv)
	_, _ = buffer.Write(keylength)
	_, _ = buffer.Write([]byte(request.Key))

	if action == "s" {
		value := request.Value
		valueLength := make([]byte, FRAGMENT_HEADER)
		var valueLengthv = uint32(len(*value))
		binary.LittleEndian.PutUint32(valueLength, valueLengthv)
		_, _ = buffer.Write(valueLength)
		_, _ = buffer.Write([]byte(*value))
	}

	return *request, buffer
}

// --- READ --------------------------------------------------------------------
func Test_ReadSetRequestFromBuffer_OK(t *testing.T) {

	// --- Setup -----------------------------------------------------------------
	myValue := "myvalue"
	request, buffer := defaultReadSetup("s", &myValue)

	// --- Execute ---------------------------------------------------------------
	parsedRequest, err := ReadFromBuffer(buffer)

	// --- Assert ----------------------------------------------------------------
	if err != nil {
		t.Fatal(err)
	}
	parsedLength := parsedRequest.TotalLength()
	reqLength := request.TotalLength()

	if parsedLength != reqLength {
		t.Fatalf("expected length '%v' but got '%v' instead", parsedLength, reqLength)
	}

}

func Test_ReadRequestFromEmptyBuffer_Errors(t *testing.T) {

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

func Test_ReadRequestFromBuffer_MovesCursor(t *testing.T) {

	myValue := "myvalue"
	// --- Setup -----------------------------------------------------------------
	firstreq, buffer := defaultReadSetup("s", &myValue)
	secondreq, secondbuffer := defaultReadSetup("d", &myValue)

	_, _ = buffer.Write(secondbuffer.Bytes())

	// --- Execute ---------------------------------------------------------------
	firstResponse, err := ReadFromBuffer(buffer)
	if err != nil {
		t.Fatal(err)
	}
	secondResponse, err := ReadFromBuffer(buffer)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ReadFromBuffer(buffer) // This one triggers an EOF.

	// --- Assert ----------------------------------------------------------------

	if firstResponse.Key != firstreq.Key {
		t.Fatalf("expected first key '%s' but got '%s' instead", firstreq.Key, firstResponse.Key)
	}

	if secondResponse.Key != secondreq.Key {
		t.Fatalf("expected second key '%s' but got '%s' instead", secondreq.Key, secondResponse.Key)
	}

	if err.Error() != "EOF" {
		t.Fatalf("expected 'EOF' as error but got '%'", err)
	}

}

// --- APPEND ------------------------------------------------------------------
func Test_AppendToNilBuffer_OK(t *testing.T) {
	myValue := "myvalue"
	req, setupBuffer := defaultReadSetup("s", &myValue)

	// --- Execute ---------------------------------------------------------------
	writtenBuffer, err := AppendToBuffer(&req, nil)

	// --- Assert ----------------------------------------------------------------
	if err != nil {
		t.Fatal(err)
	}

	if setupBuffer.Bytes()[0] != writtenBuffer.Bytes()[0] {
		t.Fatalf("written buffer does not contain the correct data")
	}
	if setupBuffer.Len() != writtenBuffer.Len() {
		t.Fatalf("written buffer does not have the expected length")
	}

}

func Test_AppendToExistingBuffer_OK(t *testing.T) {

	myValue := "myvalue"
	req, setupBuffer := defaultReadSetup("s", &myValue)
	// Maintains the original bytes length.
	comparisonBuffer := setupBuffer.Bytes()

	// --- Execute ---------------------------------------------------------------
	// Write the same request twice to double its length.
	buffer, err := AppendToBuffer(&req, setupBuffer)

	// --- Assert ----------------------------------------------------------------
	if err != nil {
		t.Fatal(err)
	}

	if len(comparisonBuffer)*2 != buffer.Len() {
		t.Fatalf("the written buffer does not have the expected length")
	}

}

func Test_AppendToBuffer_SizeLimit_Errors(t *testing.T) {

	const msg string = "hello"
	var myValue string
	for i := uint32(0); i < MESSAGE_MAX_SIZE; i++ {
		myValue = msg + myValue
	}

	req, _ := defaultReadSetup("s", &myValue)

	// --- Execute ---------------------------------------------------------------
	_, err := AppendToBuffer(&req, nil)

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

	myValue := "myvalue"
	// --- Setup -----------------------------------------------------------------
	firstreq, buffer := defaultReadSetup("s", &myValue)
	secondreq, secondbuffer := defaultReadSetup("d", &myValue)

	_, _ = buffer.Write(secondbuffer.Bytes())

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

	if firstResponse.Key != firstreq.Key {
		t.Fatalf("expected key '%s' but got '%s' instead", firstreq.Key, firstResponse.Key)
	}

	if secondResponse.Key != secondreq.Key {
		t.Fatalf("expected key '%s' but got '%s' instead", secondreq.Key, secondResponse.Key)
	}

}
