package pkg

import (
	"sync"
)

// TODO: AI(Joel): should size the array depending on the algo size in order to support more algo
var buffers = sync.Pool{
	New: func() interface{} {
		// 2 times 256 as it's only used to concat two hashes of 256 bits together
		b := make([]byte, 256+256)
		return &BuffCloser{arr: b}
	},
}

// GetConcatBuffers returns an instance of BufferCloser
func GetConcatBuffers() *BuffCloser {
	return buffers.Get().(*BuffCloser)
}

// BuffCloser represents an array of bytes
type BuffCloser struct {
	arr []byte
}

// Close puts the buffer back into the pool
func (b *BuffCloser) Close() error {
	if b != nil && b.arr != nil {
		buffers.Put(b)
	}
	return nil
}
