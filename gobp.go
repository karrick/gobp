package gobp

import (
	"bytes"
	"sync"
)

// Pool is a free-list of
//
//	p := &gobp.Pool{
//		BufSizeInit: bufSize,
//		PoolSizeMax: poolSize,
//	}
//
//      // optionally, pre-initialize buffers
//	for i := 0; i < poolSize; i++ {
//		p.Put(newBuf())
//	}
//
//      buf := p.Get()
//      buf.Reset()
//      p.Put(buf)
type Pool struct {
	// BufSizeInit is the initial byte size of newly created new buffers. When omitted, new
	// buffers have the default size of a newly created empty `bytes.Buffer` instance.
	BufSizeInit int

	// BufSizeMax is the maximum capacity of buffers allowed to be returned to the pool. Buffers
	// whose capacity is larger than this value will be released to GC.
	BufSizeMax int

	// PoolSizeMax is the maximum number of buffers the pool will hold onto. Additional buffers
	// returned to the pool will be released to GC.
	PoolSizeMax int

	free []*bytes.Buffer
	lock sync.Mutex
}

// Get acquires and returns an item from the pool. Get does not block waiting for a buffer; if the
// pool is empty a new buffer will be created and returned.
func (p *Pool) Get() *bytes.Buffer {
	p.lock.Lock()

	if len(p.free) == 0 {
		p.lock.Unlock()
		if p.BufSizeInit == 0 {
			return &bytes.Buffer{}
		}
		return bytes.NewBuffer(make([]byte, 0, p.BufSizeInit))
	}

	var bb *bytes.Buffer
	bb, p.free = p.free[0], p.free[1:]

	p.lock.Unlock()
	return bb
}

// Put will release a buffer back to the pool. If BufSizeMax is greater than 0 and the buffer's
// capacity is greater than BufSizeMax, then the buffer is released to runtime GC. If PoolSizeMax is
// greater than 0 and there are already PoolSizeMax elements in the pool, then the buffer is
// released to runtime GC. Put will not block; if the pool is full the returned buffer will be
// immediately released to runtime GC.
func (p *Pool) Put(bb *bytes.Buffer) {
	if p.BufSizeMax > 0 && bb.Cap() > p.BufSizeMax {
		return // drop buffer
	}

	p.lock.Lock()

	if p.PoolSizeMax > 0 && len(p.free) == p.PoolSizeMax {
		p.lock.Unlock()
		return // drop buffer
	}

	// store item in pool
	bb.Reset()
	p.free = append(p.free, bb)
	p.lock.Unlock()
}
