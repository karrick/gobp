package gobp_test

import (
	"bytes"
	"fmt"
	"sync"
	"testing"

	"github.com/karrick/gobp"
)

const count = 1000

func TestGobp2Stress(t *testing.T) {
	const bufSize = 16 * 1024
	const poolSizeMax = 8
	const poolSizeMin = poolSizeMax / 2
	const goroutines = poolSizeMax * 2
	const perGoRoutineIterations = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	pool := &gobp.Pool2{
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

func BenchmarkQueue(b *testing.B) {
	var v int

	slice := make([]int, count)
	for i := 0; i < count; i++ {
		slice[i] = i
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		v, slice = slice[0], slice[1:] // shift
		slice = append(slice, v)       // unshift
	}
}

func BenchmarkStack(b *testing.B) {
	var v int

	slice := make([]int, count)
	for i := 0; i < count; i++ {
		slice[i] = i
	}
	b.ResetTimer()

	const cmo = count - 1

	for i := 0; i < b.N; i++ {
		v, slice = slice[cmo], slice[:cmo] // pop
		slice = append(slice, v)           // push
	}
}
