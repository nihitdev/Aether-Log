package protocol

import (
	"bytes"
	"testing"
)

func TestFrameReadWrite(t *testing.T) {
	payload := []byte("hello world")
	frame := &Frame{
		Type:    TypeData,
		Payload: payload,
	}

	var buf bytes.Buffer
	err := WriteFrame(&buf, frame)
	if err != nil {
		t.Fatalf("WriteFrame failed: %v", err)
	}

	readBuf := make([]byte, 1024)
	readFrame, err := ReadFrame(&buf, readBuf)
	if err != nil {
		t.Fatalf("ReadFrame failed: %v", err)
	}

	if readFrame.Type != frame.Type {
		t.Errorf("Expected type %v, got %v", frame.Type, readFrame.Type)
	}

	if !bytes.Equal(readFrame.Payload, frame.Payload) {
		t.Errorf("Expected payload %s, got %s", frame.Payload, readFrame.Payload)
	}
}

func TestInvalidMagic(t *testing.T) {
	data := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00}
	reader := bytes.NewReader(data)
	_, err := ReadFrame(reader, nil)
	if err != ErrInvalidMagic {
		t.Errorf("Expected ErrInvalidMagic, got %v", err)
	}
}
