package storage

import (
	"fmt"
	"os"
	"sync"
)

// Rotator serializes writes and rotates before a write would exceed MaxBytes.
// A single oversized write is kept intact. Retain counts archived files.
type Rotator struct {
	mu       sync.Mutex
	filename string
	file     *os.File
	size     int64
	maxBytes int64
	retain   int
}

func NewRotator(filename string) (*Rotator, error) { return Open(filename, 0, 0) }
func Open(filename string, maxBytes int64, retain int) (*Rotator, error) {
	if maxBytes < 0 || retain < 0 {
		return nil, fmt.Errorf("rotation limits must be nonnegative")
	}
	r := &Rotator{filename: filename, maxBytes: maxBytes, retain: retain}
	if err := r.open(); err != nil {
		return nil, err
	}
	return r, nil
}
func (r *Rotator) open() error {
	f, err := os.OpenFile(r.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	r.file = f
	r.size = info.Size()
	return nil
}
func (r *Rotator) rotate() error {
	if err := r.file.Close(); err != nil {
		return err
	}
	r.file = nil
	for i := r.retain; i >= 1; i-- {
		src := r.filename
		if i > 1 {
			src = fmt.Sprintf("%s.%d", r.filename, i-1)
		}
		dst := fmt.Sprintf("%s.%d", r.filename, i)
		if err := os.Rename(src, dst); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("rotate %s: %w", src, err)
		}
	}
	if r.retain == 0 {
		if err := os.Remove(r.filename); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return r.open()
}
func (r *Rotator) Write(data []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file == nil {
		return 0, os.ErrClosed
	}
	if r.maxBytes > 0 && r.size > 0 && int64(len(data)) > r.maxBytes-r.size {
		if err := r.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := r.file.Write(data)
	r.size += int64(n)
	return n, err
}
func (r *Rotator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file == nil {
		return nil
	}
	err := r.file.Close()
	r.file = nil
	return err
}
