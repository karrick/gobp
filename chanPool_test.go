package gobp_test

import (
	"bytes"
	"testing"
)

// ChanPool is a buffer free-list.
//
//	p := &gobp.NewChanPool(&gobp.ChanPool{
//		BufSizeInit: bufSize,
//		PoolSizeMax: poolSize,
//	})
//
//      buf := p.Get()
//      buf.Reset()
//      p.Put(buf)
type ChanPool struct {
	// BufSizeMax is the maximum capacity of buffers allowed to be returned to the pool. Buffers
	// whose capacity is larger than this value will be released to GC.
	BufSizeMax int

	// PoolSizeMax is the maximum number of buffers the pool will hold onto. Additional buffers
	// returned to the pool will be released to GC.
	PoolSizeMax int

	ch chan *bytes.Buffer
}

func NewChanPool(cfg *ChanPool) *ChanPool {
	return &ChanPool{
		BufSizeMax: cfg.BufSizeMax,
		ch:         make(chan *bytes.Buffer, cfg.PoolSizeMax),
	}
}

// Get acquires and returns an item from the pool. Get does not block waiting for a buffer; if the
// pool is empty a new buffer will be created and returned.
func (p *ChanPool) Get() *bytes.Buffer {
	select {
	case b := <-p.ch:
		return b
	default:
		// if p.BufSizeInit == 0 {
		// 	return &bytes.Buffer{}
		// }
		// return bytes.NewBuffer(make([]byte, 0, p.BufSizeInit))
		return bytes.NewBuffer(make([]byte, 0, p.BufSizeMax))
	}
}

// Put will release a buffer back to the pool. If BufSizeMax is greater than 0 and the buffer's
// capacity is greater than BufSizeMax, then the buffer is released to runtime GC. If PoolSizeMax is
// greater than 0 and there are already PoolSizeMax elements in the pool, then the buffer is
// released to runtime GC. Put will not block; if the pool is full the returned buffer will be
// immediately released to runtime GC.
func (p *ChanPool) Put(bb *bytes.Buffer) {
	if p.BufSizeMax > 0 && bb.Cap() > p.BufSizeMax {
		return // drop buffer
	}
	bb.Reset()
	select {
	case p.ch <- bb:
	default:
	}
}

////////////////////////////////////////

func newGobpChan() pool {
	p := NewChanPool(&ChanPool{
		// BufSizeInit: bufSize,
		BufSizeMax:  bufSize,
		PoolSizeMax: poolSize,
	})
	for i := 0; i < poolSize; i++ {
		p.Put(newBuf())
	}
	return p
}

func BenchmarkGobpChan(b *testing.B) {
	benchLowAndHigh(b, newGobpChan())
}
