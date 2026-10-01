package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
	"time"
)

func TestFrameRoundTrip(t *testing.T) {
	var wire bytes.Buffer

	want := []byte("hello aether")

	if err := WriteFrame(&wire, TypeData, want); err != nil {
		t.Fatalf("WriteFrame: %v", err)
	}

	frame, err := ReadFrame(&wire, 1024)
	if err != nil {
		t.Fatalf("ReadFrame: %v", err)
	}

	if frame.Type != TypeData {
		t.Fatalf("type = %d, want %d", frame.Type, TypeData)
	}

	if !bytes.Equal(frame.Payload, want) {
		t.Fatalf("payload = %q, want %q", frame.Payload, want)
	}
}

func TestRejectInvalidMagic(t *testing.T) {
	header := make([]byte, HeaderSize)

	binary.BigEndian.PutUint32(header[0:4], 0xDEADBEEF)
	binary.BigEndian.PutUint16(header[4:6], Version)
	header[6] = byte(TypeData)

	_, err := ReadFrame(bytes.NewReader(header), 1024)

	if !errors.Is(err, ErrInvalidMagic) {
		t.Fatalf("got %v, want ErrInvalidMagic", err)
	}
}

func TestRejectOversizedPayload(t *testing.T) {
	header := make([]byte, HeaderSize)

	binary.BigEndian.PutUint32(header[0:4], Magic)
	binary.BigEndian.PutUint16(header[4:6], Version)
	header[6] = byte(TypeData)
	binary.BigEndian.PutUint32(header[7:11], 4096)

	_, err := ReadFrame(bytes.NewReader(header), 128)

	if !errors.Is(err, ErrPayloadTooLarge) {
		t.Fatalf("got %v, want ErrPayloadTooLarge", err)
	}
}

func TestSlowDownRoundTrip(t *testing.T) {
	want := 150 * time.Millisecond

	got, err := ParseSlowDown(SlowDownPayload(want))
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestMalformedFrames(t *testing.T) {
	var wire bytes.Buffer
	WriteFrame(&wire, TypeData, []byte("abc"))
	valid := wire.Bytes()
	for _, tc := range []struct {
		name   string
		mutate func([]byte)
		want   error
	}{
		{"version", func(p []byte) { p[5] = 2 }, ErrInvalidVersion},
		{"type", func(p []byte) { p[6] = 99 }, ErrInvalidType},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := append([]byte(nil), valid...)
			tc.mutate(p)
			_, err := ReadFrame(bytes.NewReader(p), 10)
			if !errors.Is(err, tc.want) {
				t.Fatal(err)
			}
		})
	}
	for n := 1; n < len(valid); n++ {
		if _, err := ReadFrame(bytes.NewReader(valid[:n]), 10); err == nil {
			t.Fatalf("accepted truncated length %d", n)
		}
	}
	if err := WriteFrame(&bytes.Buffer{}, 99, nil); !errors.Is(err, ErrInvalidType) {
		t.Fatal(err)
	}
	if _, err := ParseSlowDown(nil); err == nil {
		t.Fatal("invalid slowdown")
	}
}
func BenchmarkEncode(b *testing.B) {
	p := []byte("example record")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		WriteFrame(io.Discard, TypeData, p)
	}
}
func BenchmarkDecode(b *testing.B) {
	var wire bytes.Buffer
	WriteFrame(&wire, TypeData, []byte("example record"))
	p := wire.Bytes()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ReadFrame(bytes.NewReader(p), 1024)
	}
}

type shortWriter struct{ bytes.Buffer }

func (w *shortWriter) Write(p []byte) (int, error) {
	if len(p) > 2 {
		p = p[:2]
	}
	return w.Buffer.Write(p)
}
func TestShortWrites(t *testing.T) {
	w := &shortWriter{}
	if err := WriteFrame(w, TypeData, []byte("abc")); err != nil {
		t.Fatal(err)
	}
	f, err := ReadFrame(&w.Buffer, 3)
	if err != nil || string(f.Payload) != "abc" {
		t.Fatal(f, err)
	}
}

type zeroWriter struct{}

func (zeroWriter) Write([]byte) (int, error) { return 0, nil }
func TestZeroWrite(t *testing.T) {
	if !errors.Is(WriteFrame(zeroWriter{}, TypeData, nil), io.ErrShortWrite) {
		t.Fatal("zero write not rejected")
	}
}
