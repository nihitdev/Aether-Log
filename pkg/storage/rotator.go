package storage

import (
	"os"
	"sync"
)

type Rotator struct {
	mu       sync.Mutex
	filename string
	file     *os.File
}

func NewRotator(filename string) (*Rotator, error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &Rotator{
		filename: filename,
		file:     f,
	}, nil
}

func (r *Rotator) Write(data []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// In a production system, we would check file size here and rotate
	return r.file.Write(data)
}

func (r *Rotator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.file.Close()
}
