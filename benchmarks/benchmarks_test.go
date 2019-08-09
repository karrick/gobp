package benchmarks

import (
	"bytes"
	"runtime"
	"sync"
	"testing"

	"github.com/karrick/gobp"
)

const (
	bufSize  = 32 * 1024 // based on Go allocation slab size
	poolSize = 64
)

func newBuf() *bytes.Buffer { return bytes.NewBuffer(make([]byte, 0, bufSize)) }

type pool interface {
	Get() *bytes.Buffer
	Put(*bytes.Buffer)
}

func newGobp() pool {
	p := &gobp.Pool{
		BufSizeInit: bufSize,
		PoolSizeMax: poolSize,
	}
	for i := 0; i < poolSize; i++ {
		p.Put(newBuf())
	}
	return p
}

////////////////////////////////////////
// NoPool

type noPool struct{}

func (p *noPool) Get() *bytes.Buffer {
	return newBuf()
}

func (p *noPool) Put(_ *bytes.Buffer) {
}

////////////////////////////////////////
// sync.Pool

type syncPool struct {
	p *sync.Pool
}

func (p *syncPool) Get() *bytes.Buffer {
	return p.p.Get().(*bytes.Buffer)
}

func (p *syncPool) Put(b *bytes.Buffer) {
	p.p.Put(b)
}

func newSyncPool() pool {
	p := &syncPool{
		p: &sync.Pool{
			New: func() interface{} { return newBuf() },
		},
	}
	for i := 0; i < poolSize; i++ {
		p.p.Put(newBuf())
	}
	return p
}

////////////////////////////////////////

func exercise(p pool) {
	buf := p.Get()
	defer p.Put(buf)

	for i := 0; i < bufSize; i++ {
		buf.WriteByte(byte(i % 256))
	}
}

func benchLowAndHigh(b *testing.B, p pool) {
	b.Run("Low", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				exercise(p)
			}
		})
	})
	b.Run("High", func(b *testing.B) {
		concurrency := runtime.NumCPU() * 100
		var wg sync.WaitGroup
		wg.Add(concurrency)

		b.ResetTimer()

		for c := 0; c < concurrency; c++ {
			go func() {
				defer wg.Done()

				for n := 0; n < b.N; n++ {
					exercise(p)
				}
			}()
		}

		wg.Wait()
	})
}

////////////////////////////////////////

func BenchmarkGobp(b *testing.B) {
	benchLowAndHigh(b, newGobp())
}

func BenchmarkNoPool(b *testing.B) {
	benchLowAndHigh(b, new(noPool))
}

func BenchmarkSyncPool(b *testing.B) {
	benchLowAndHigh(b, newSyncPool())
}
