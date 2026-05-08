package protocol

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	MagicNumber = 0xAE744552
	Version     = 1
)

type FrameType uint8

const (
	TypeData     FrameType = 1
	TypeSlowDown FrameType = 2
)

var (
	ErrInvalidMagic   = errors.New("invalid magic number")
	ErrInvalidVersion = errors.New("unsupported protocol version")
)

// Frame represents a single message in the Aether-Log protocol.
type Frame struct {
	Type    FrameType
	Payload []byte
}

// Header: Magic(4) + Version(2) + Type(1) + Length(4) = 11 bytes
const HeaderSize = 11

// WriteFrame writes a frame to the provided writer.
func WriteFrame(w io.Writer, f *Frame) error {
	header := make([]byte, HeaderSize)
	binary.BigEndian.PutUint32(header[0:4], MagicNumber)
	binary.BigEndian.PutUint16(header[4:6], Version)
	header[6] = uint8(f.Type)
	binary.BigEndian.PutUint32(header[7:11], uint32(len(f.Payload)))

	if _, err := w.Write(header); err != nil {
		return err
	}
	if len(f.Payload) > 0 {
		_, err := w.Write(f.Payload)
		return err
	}
	return nil
}

// ReadFrame reads a frame from the provided reader.
// It uses the provided buffer to read the payload if possible, otherwise it allocates.
// This is to support the zero-allocation goal.
func ReadFrame(r io.Reader, buf []byte) (*Frame, error) {
	header := make([]byte, HeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	magic := binary.BigEndian.Uint32(header[0:4])
	if magic != MagicNumber {
		return nil, ErrInvalidMagic
	}

	version := binary.BigEndian.Uint16(header[4:6])
	if version != Version {
		return nil, ErrInvalidVersion
	}

	fType := FrameType(header[6])
	length := binary.BigEndian.Uint32(header[7:11])

	var payload []byte
	if length > 0 {
		if uint32(len(buf)) >= length {
			payload = buf[:length]
		} else {
			payload = make([]byte, length)
		}
		if _, err := io.ReadFull(r, payload); err != nil {
			return nil, err
		}
	}

	return &Frame{
		Type:    fType,
		Payload: payload,
	}, nil
}
