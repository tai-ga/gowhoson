package whoson

import (
	"sync"
)

type Buffer struct {
	pool  *BufferPool
	buf   []byte
	count int
}

func (b *Buffer) Free() {
	b.pool.put(b)
}

type BufferPool struct {
	p *sync.Pool
}

func NewBufferPool() *BufferPool {
	return &BufferPool{
		p: &sync.Pool{
			New: func() interface{} {
				return &Buffer{buf: make([]byte, udpByteSize)}
			},
		},
	}
}

func (p *BufferPool) Get() *Buffer {
	b := p.p.Get().(*Buffer)
	b.pool = p
	return b
}

func (p *BufferPool) put(b *Buffer) {
	p.p.Put(b)
}
