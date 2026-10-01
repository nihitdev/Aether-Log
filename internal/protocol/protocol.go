package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"time"
)

const (
	Magic      uint32 = 0xAE744552
	Version    uint16 = 1
	HeaderSize        = 11
)

type FrameType uint8

const (
	TypeData     FrameType = 1
	TypeSlowDown FrameType = 2
)

var (
	ErrInvalidMagic    = errors.New("invalid magic number")
	ErrInvalidVersion  = errors.New("unsupported protocol version")
	ErrInvalidType     = errors.New("unsupported frame type")
	ErrPayloadTooLarge = errors.New("payload too large")
)

type Frame struct {
	Type    FrameType
	Payload []byte
}

func validType(t FrameType) bool {
	return t == TypeData || t == TypeSlowDown
}

func WriteFrame(w io.Writer, frameType FrameType, payload []byte) error {
	if !validType(frameType) {
		return ErrInvalidType
	}

	if uint64(len(payload)) > math.MaxUint32 {
		return ErrPayloadTooLarge
	}

	header := make([]byte, HeaderSize)

	binary.BigEndian.PutUint32(header[0:4], Magic)
	binary.BigEndian.PutUint16(header[4:6], Version)
	header[6] = byte(frameType)
	binary.BigEndian.PutUint32(header[7:11], uint32(len(payload)))

	if err := writeAll(w, header); err != nil {
		return err
	}

	return writeAll(w, payload)
}

func ReadFrame(r io.Reader, maxPayload uint32) (Frame, error) {
	header := make([]byte, HeaderSize)

	if _, err := io.ReadFull(r, header); err != nil {
		return Frame{}, err
	}

	if binary.BigEndian.Uint32(header[0:4]) != Magic {
		return Frame{}, ErrInvalidMagic
	}

	if binary.BigEndian.Uint16(header[4:6]) != Version {
		return Frame{}, ErrInvalidVersion
	}

	frameType := FrameType(header[6])

	if !validType(frameType) {
		return Frame{}, ErrInvalidType
	}

	length := binary.BigEndian.Uint32(header[7:11])

	if length > maxPayload {
		return Frame{}, fmt.Errorf(
			"%w: %d > %d",
			ErrPayloadTooLarge,
			length,
			maxPayload,
		)
	}

	payload := make([]byte, length)

	if length > 0 {
		if _, err := io.ReadFull(r, payload); err != nil {
			return Frame{}, err
		}
	}

	return Frame{
		Type:    frameType,
		Payload: payload,
	}, nil
}

func SlowDownPayload(delay time.Duration) []byte {
	ms := delay.Milliseconds()

	if ms < 0 {
		ms = 0
	}

	if ms > math.MaxUint32 {
		ms = math.MaxUint32
	}

	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(ms))

	return payload
}

func ParseSlowDown(payload []byte) (time.Duration, error) {
	if len(payload) != 4 {
		return 0, fmt.Errorf(
			"slowdown payload must contain 4 bytes, got %d",
			len(payload),
		)
	}

	ms := binary.BigEndian.Uint32(payload)

	return time.Duration(ms) * time.Millisecond, nil
}

func writeAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}

		if n == 0 {
			return io.ErrShortWrite
		}

		data = data[n:]
	}

	return nil
}
