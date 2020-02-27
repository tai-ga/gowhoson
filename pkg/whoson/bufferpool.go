package whoson

import (
	"sync"
)

// Buffer hold information for buffer pool.
type Buffer struct {
	pool  *BufferPool
	buf   []byte
	count int
}

// Free push to buffer pool.
func (b *Buffer) Free() {
	b.pool.put(b)
}

// BufferPool hold information for sync.Pool.
type BufferPool struct {
	p *sync.Pool
}

// NewBufferPool return new BufferPool struct pointer.
func NewBufferPool() *BufferPool {
	return &BufferPool{
		p: &sync.Pool{
			New: func() interface{} {
				return &Buffer{buf: make([]byte, udpByteSize)}
			},
		},
	}
}

// Get get from buffer pool.
func (p *BufferPool) Get() *Buffer {
	b := p.p.Get().(*Buffer)
	b.pool = p
	return b
}

func (p *BufferPool) put(b *Buffer) {
	p.p.Put(b)
}
