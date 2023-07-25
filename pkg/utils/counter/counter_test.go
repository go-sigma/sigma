// Copyright 2023 sigma
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package counter

import (
	"crypto/rand"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/utils/hash"
)

func TestCounter(t *testing.T) {
	type args struct {
		do func(int64)
	}
	tests := []struct {
		name            string
		prepareFunc     func() (io.Reader, string)
		prepareCallback func(*int64) func(int64)
		args            args
		expectFunc      func(*testing.T, *Counter, string, *int64)
		wantErr         bool
	}{
		{
			name: "basic",
			prepareFunc: func() (io.Reader, string) {
				testString := "hello"
				reader := strings.NewReader(testString)
				h, _ := hash.String(testString)
				return reader, h
			},
			args: args{
				do: func(n int64) {},
			},
			expectFunc: func(t *testing.T, r *Counter, s string, _ *int64) {
				data, err := io.ReadAll(r)
				assert.NoError(t, err)
				h, err := hash.String(string(data))
				assert.NoError(t, err)
				assert.Equal(t, s, h)
			},
			wantErr: false,
		},
		{
			name: "big file",
			prepareFunc: func() (io.Reader, string) {
				var bigFile = "test-big-file.bin"
				file, _ := os.Create(bigFile)
				for i := 0; i < 100; i++ { // 100M
					data := make([]byte, 1<<20)
					_, _ = rand.Read(data)
					_, _ = file.Write(data)
				}
				_ = file.Close()
				s, _ := hash.File(bigFile)
				r, _ := os.Open(bigFile)
				return r, s
			},
			args: args{
				do: func(n int64) {},
			},
			expectFunc: func(t *testing.T, r *Counter, s string, _ *int64) {
				var bigFile = "test-big-file-receive.bin"
				file, err := os.Create(bigFile)
				assert.NoError(t, err)
				_, err = io.Copy(file, r)
				assert.NoError(t, err)
				err = file.Close()
				assert.NoError(t, err)
				h, err := hash.File(bigFile)
				assert.NoError(t, err)
				assert.Equal(t, s, h)
				err = os.Remove(bigFile)
				assert.NoError(t, err)
				err = os.RemoveAll("test-big-file.bin")
				assert.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name: "ticker",
			prepareFunc: func() (io.Reader, string) {
				var bigFile = "test-ticker.bin"
				file, _ := os.Create(bigFile)
				for i := 0; i < 1000; i++ { // 100M
					data := make([]byte, 1<<20)
					_, _ = rand.Read(data)
					_, _ = file.Write(data)
				}
				_ = file.Close()
				s, _ := hash.File(bigFile)
				r, _ := os.Open(bigFile)
				return r, s
			},
			prepareCallback: func(n *int64) func(int64) {
				return func(i int64) {
					atomic.SwapInt64(n, i)
				}
			},
			args: args{},
			expectFunc: func(t *testing.T, r *Counter, s string, n *int64) {
				var bigFile = "test-ticker-receive.bin"
				file, err := os.Create(bigFile)
				assert.NoError(t, err)
				_, err = io.Copy(file, r)
				assert.NoError(t, err)
				err = file.Close()
				assert.NoError(t, err)
				h, err := hash.File(bigFile)
				assert.NoError(t, err)
				assert.Equal(t, s, h)
				err = os.Remove(bigFile)
				assert.NoError(t, err)
				err = os.RemoveAll("test-ticker.bin")
				assert.NoError(t, err)
				assert.Equal(t, int64(1000*1<<20), atomic.LoadInt64(n))
			},
			wantErr: false,
		},
		{
			name: "ticker with panic",
			prepareFunc: func() (io.Reader, string) {
				var bigFile = "test-ticker.bin"
				file, _ := os.Create(bigFile)
				for i := 0; i < 1000; i++ { // 100M
					data := make([]byte, 1<<20)
					_, _ = rand.Read(data)
					_, _ = file.Write(data)
				}
				_ = file.Close()
				s, _ := hash.File(bigFile)
				r, _ := os.Open(bigFile)
				return r, s
			},
			prepareCallback: func(n *int64) func(int64) {
				return func(i int64) {
					atomic.SwapInt64(n, i)
					panic("panic")
				}
			},
			args: args{},
			expectFunc: func(t *testing.T, r *Counter, s string, n *int64) {
				var bigFile = "test-ticker-receive.bin"
				file, err := os.Create(bigFile)
				assert.NoError(t, err)
				_, err = io.Copy(file, r)
				assert.NoError(t, err)
				err = file.Close()
				assert.NoError(t, err)
				h, err := hash.File(bigFile)
				assert.NoError(t, err)
				assert.Equal(t, s, h)
				err = os.Remove(bigFile)
				assert.NoError(t, err)
				err = os.RemoveAll("test-ticker.bin")
				assert.NoError(t, err)
				assert.Equal(t, int64(1000*1<<20), r.Count())
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, s := tt.prepareFunc()
			n := int64(0)
			if tt.prepareCallback != nil {
				tt.args.do = tt.prepareCallback(&n)
			}
			counter := NewCounter(r)
			counter.Tick(tt.args.do, time.NewTicker(time.Millisecond*100))
			tt.expectFunc(t, counter, s, &n)
		})
	}
}
