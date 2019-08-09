package gobp_test

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/karrick/gobp"
)

const (
	poolSize = 64
	bufSize  = 32 * 1024
)

func TestGobpStress(t *testing.T) {
	const poolSizeMax = 8
	const poolSizeMin = poolSizeMax / 2
	const goroutines = poolSizeMax * 2
	const perGoRoutineIterations = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	pool := &gobp.Pool{
		BufSizeInit: bufSize,
		BufSizeMax:  bufSize << 1,
		PoolSizeMax: poolSizeMax,
	}

	// optionally fill the pool with pre-allocated buffers
	for i := 0; i < poolSizeMin; i++ {
		pool.Put(bytes.NewBuffer(make([]byte, 0, bufSize)))
	}

	// run some concurrency tests
	for c := 0; c < goroutines; c++ {
		go func() {
			defer wg.Done()

			for i := 0; i < perGoRoutineIterations; i++ {
				if err := grabBufferAndUseIt(pool); err != nil {
					fmt.Println(err)
				}
			}
		}()
	}

	wg.Wait()
}

func grabBufferAndUseIt(pool *gobp.Pool) error {
	// NOTE: Like all resources obtained from a pool, failing to release results in resource
	// leaks.
	bb := pool.Get()
	defer pool.Put(bb)

	extra := rand.Intn(bufSize) - bufSize/2 // 4096 +/- 2048

	for i := 0; i < extra+bufSize; i++ {
		if rand.Intn(100000000) == 1 {
			return errors.New("random error to illustrate need to return resource to pool")
		}
		bb.WriteByte(byte(i % 256))
	}
	return nil
}
