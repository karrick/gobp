# gobp

Go library for using a pool, also known as a free-list, of byte buffers.

## Background

When a program is going to repeatedly allocate memory, use the memory, and finally free the memory,
it creates additional stress on the runtime memory manager and garbage collector. This stress is
often mitigated in programs by using a free-list of memory blocks by the program.

However, managing such a free-list can adversely impact performance in a concurrent environment. The
amount of impact is dependent upon the number of concurrent accesses to the free-list, along with
the specific concurrency primitives used to manage the free-list.

Obvious the most performant algorithm is to _not_ use a free-list of buffers and to allocate a new
buffer when needed, then release the buffer when no longer needed. This places the entire burden of
memory management on the runtime memory manager and garbage collector, and for some applications,
might be undesirable.

A few excellent articles were published that resurfaced the topic of using free-lists of
`byte.Buffer` structures in Go.  Because the Go runtime includes facilities to manage free-lists, I
was curious about the performance characteristics of various methods of achieving this goal, and
decided to benchmark these options.

* https://blog.cloudflare.com/recycling-memory-buffers-in-go/
* https://elithrar.github.io/article/using-buffer-pools-with-go/

## Description

This library provides a pool of buffers, namely `*bytes.Buffer` instances, and attempts to strike a
balance between relying on the runtime memory manager and lock-free concurrency. Note that lock-free
is in contrast to a wait-free algorithm.

## Usage

### Simple Example

Simple example that illustrates basic pool creation and use. Note that when no `BufSizeInit` is
specified, the pool will create a new `bytes.Buffer` with whatever default _that_ uses.

```Go
	p := new(gobp.Pool)

	buf := p.Get()
	buf.WriteString("test")
	p.Put(buf)
```

### Concurrent Example

Here is a more complex example illustrating concurrent access to the pool.

```Go
	package main

	import (
		"bytes"
		"errors"
		"fmt"
		"math/rand"
		"sync"

		"github.com/karrick/gobp"
	)

	const (
		bufSize                = 16 * 1024
		poolSize               = 50
		perGoRoutineIterations = 100
	)

	func main() {
		const goroutines = poolSize

		var wg sync.WaitGroup
		wg.Add(goroutines)

		pool := &gobp.Pool{
			BufSizeInit: bufSize,
			BufSizeMax:  bufSize + 1024,
			PoolSizeMax: poolSize,
		}

		// optionally fill the pool with pre-allocated buffers
		for i := 0; i < poolSize; i++ {
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
		// NOTE: Like all resources obtained from a pool, failing to release
		// results in resource leaks.
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
```

## Development Note

Build tags have been included to protect against including additional libraries used only for
benchmarks. To run comparison benchmarks, add the `bench` tag to the command line as demonstrated
below.

```Bash
go test -v -bench=. -tags=bench
```
