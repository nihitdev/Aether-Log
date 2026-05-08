package buffer

import "sync"

const DefaultBufferSize = 32 * 1024 // 32KB

var pool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, DefaultBufferSize)
		return &b
	},
}

// Get returns a buffer from the pool.
func Get() *[]byte {
	return pool.Get().(*[]byte)
}

// Put returns a buffer to the pool.
func Put(b *[]byte) {
	if cap(*b) < DefaultBufferSize {
		return
	}
	*b = (*b)[:DefaultBufferSize]
	pool.Put(b)
}
