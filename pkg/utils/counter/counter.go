package counter

import (
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

// Counter is a counter for the number of bytes read from the reader
type Counter struct {
	io.Reader
	count int64
}

// NewCounter function for create new Counter
func NewCounter(r io.Reader) *Counter {
	return &Counter{
		Reader: r,
	}
}

// Read implements io.Reader interface
func (c *Counter) Read(buf []byte) (int, error) {
	n, err := c.Reader.Read(buf)

	// 有些 reader 的实现里会返回 -1
	if n >= 0 {
		atomic.AddInt64(&c.count, int64(n))
	}

	return n, err
}

// Count count the number of bytes read from the reader
func (c *Counter) Count() int64 {
	return atomic.LoadInt64(&c.count)
}

// Tick ticker for progress
// consider to use ticker.Stop() to stop the ticker by yourself
func (c *Counter) Tick(do func(int64), ticker *time.Ticker) {
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				fmt.Printf("proxy reader ticker panic: %v\n", err)
			}
		}()
		for range ticker.C {
			do(c.Count())
		}
	}()
}
