// The MIT License (MIT)
//
// Copyright © 2023 Tosone <i@tosone.cn>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
