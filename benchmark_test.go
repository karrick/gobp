package gobp_test

import (
	"bytes"
	"log"
	"runtime"
	"sync"
	"testing"

	"github.com/karrick/gobp"
	"github.com/karrick/gopool"
)

const (
	bufSize  = 16 * 1024
	poolSize = 64
)

func newBuf() *bytes.Buffer { return bytes.NewBuffer(make([]byte, 0, bufSize)) }

func newGobp() (func() *bytes.Buffer, func(*bytes.Buffer)) {
	p := &gobp.Pool{
		BufSizeInit: bufSize,
		PoolSizeMax: poolSize,
	}
	for i := 0; i < poolSize; i++ {
		p.Put(newBuf())
	}
	return p.Get, p.Put
}

func newGopool() (func() *bytes.Buffer, func(*bytes.Buffer)) {
	p, err := gopool.New(gopool.Size(poolSize),
		gopool.Factory(func() (interface{}, error) {
			return newBuf(), nil
		}),
		gopool.Reset(func(item interface{}) {
			item.(*bytes.Buffer).Reset()
		}))
	if err != nil {
		log.Fatal(err)
	}
	setup := func() *bytes.Buffer {
		return p.Get().(*bytes.Buffer)
	}
	teardown := func(buf *bytes.Buffer) {
		p.Put(buf)
	}
	return setup, teardown
}

func newSyncPool() (func() *bytes.Buffer, func(*bytes.Buffer)) {
	p := &sync.Pool{
		New: func() interface{} { return newBuf() },
	}
	for i := 0; i < poolSize; i++ {
		p.Put(newBuf())
	}
	setup := func() *bytes.Buffer {
		return p.Get().(*bytes.Buffer)
	}
	teardown := func(buf *bytes.Buffer) {
		buf.Reset()
		p.Put(buf)
	}
	return setup, teardown
}

////////////////////////////////////////

func exercise(setup func() *bytes.Buffer, teardown func(*bytes.Buffer)) {
	buf := setup()
	defer teardown(buf)

	for i := 0; i < bufSize; i++ {
		buf.WriteByte(byte(i % 256))
	}
}

func benchmarkLow(b *testing.B, setup func() *bytes.Buffer, teardown func(*bytes.Buffer)) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			exercise(setup, teardown)
		}
	})
}

func benchmarkHigh(b *testing.B, setup func() *bytes.Buffer, teardown func(*bytes.Buffer)) {
	concurrency := runtime.NumCPU() * 100
	var wg sync.WaitGroup
	wg.Add(concurrency)

	b.ResetTimer()

	for c := 0; c < concurrency; c++ {
		go func() {
			defer wg.Done()

			for n := 0; n < b.N; n++ {
				exercise(setup, teardown)
			}
		}()
	}

	wg.Wait()
}

////////////////////////////////////////
// Low Concurrency

func BenchmarkLowConcurrencyGobp(b *testing.B) {
	setup, teardown := newGobp()
	benchmarkLow(b, setup, teardown)
}

func BenchmarkLowConcurrencyGopool(b *testing.B) {
	setup, teardown := newGopool()
	benchmarkLow(b, setup, teardown)
}

func BenchmarkLowConcurrencyNoPool(b *testing.B) {
	setup, teardown := newBuf, func(_ *bytes.Buffer) {}
	benchmarkLow(b, setup, teardown)
}

func BenchmarkLowConcurrencySyncPool(b *testing.B) {
	setup, teardown := newSyncPool()
	benchmarkLow(b, setup, teardown)
}

////////////////////////////////////////
// High Concurrency

func BenchmarkHighConcurrencyGobp(b *testing.B) {
	setup, teardown := newGobp()
	benchmarkHigh(b, setup, teardown)
}

func BenchmarkHighConcurrencyGopool(b *testing.B) {
	setup, teardown := newGopool()
	benchmarkHigh(b, setup, teardown)
}

func BenchmarkHighConcurrencyNoPool(b *testing.B) {
	setup, teardown := newBuf, func(_ *bytes.Buffer) {}
	benchmarkHigh(b, setup, teardown)
}

func BenchmarkHighConcurrencySyncPool(b *testing.B) {
	setup, teardown := newSyncPool()
	benchmarkHigh(b, setup, teardown)
}
