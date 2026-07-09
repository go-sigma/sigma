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

package hash

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestHashString(t *testing.T) {
	type args struct {
		str    string
		method []string
	}
	tests := []struct {
		name     string
		args     args
		wantHash string
		wantErr  bool
	}{
		{
			name: "md5",
			args: args{
				str:    "hello",
				method: []string{"md5"},
			},
			wantHash: "5d41402abc4b2a76b9719d911017c592",
			wantErr:  false,
		},
		{
			name: "sha1",
			args: args{
				str:    "hello",
				method: []string{"sha1"},
			},
			wantHash: "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d",
			wantErr:  false,
		},
		{
			name: "sha256",
			args: args{
				str:    "hello",
				method: []string{"sha256"},
			},
			wantHash: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
			wantErr:  false,
		},
		{
			name: "sha512",
			args: args{
				str:    "hello",
				method: []string{"sha512"},
			},
			wantHash: "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHash, err := String(tt.args.str, tt.args.method...)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHash != tt.wantHash {
				t.Errorf("HashString() = %v, want %v", gotHash, tt.wantHash)
			}
		})
	}
}

func TestHashFile(t *testing.T) {
	type args struct {
		file   string
		method []string
	}
	tests := []struct {
		name     string
		args     args
		wantHash string
		wantErr  bool
	}{
		{
			name: "md5",
			args: args{
				file:   "ci.txt",
				method: []string{"md5"},
			},
			wantHash: "f621bae50c4c5099943aaaa3ef51c12b",
			wantErr:  false,
		},
		{
			name: "sha1",
			args: args{
				file:   "ci.txt",
				method: []string{"sha1"},
			},
			wantHash: "c560dd44beba55babb7e581de40861d8412d1cb6",
			wantErr:  false,
		},
		{
			name: "sha256",
			args: args{
				file:   "ci.txt",
				method: []string{"sha256"},
			},
			wantHash: "9379c18887bd8cf8895c856bfbb1d4704a13595ff0a6cf4e7823770728cd6ee2",
			wantErr:  false,
		},
		{
			name: "sha512",
			args: args{
				file:   "ci.txt",
				method: []string{"sha512"},
			},
			wantHash: "33500e835ae5009093d54ce1841a723e91a68664aa068a9d4493f1c168f67ab656792ce396de5d599e334e6d143fe8878f7371edea80e2f389511bdba7d095f4",
			wantErr:  false,
		},
		{
			name: "file not exist",
			args: args{
				file:   "not_exist.go",
				method: []string{"sha512"},
			},
			wantHash: "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHash, err := File(tt.args.file, tt.args.method...)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHash != tt.wantHash {
				t.Errorf("HashFile() = %v, want %v", gotHash, tt.wantHash)
			}
		})
	}
}

func TestHashFileVerify(t *testing.T) {
	type args struct {
		file   string
		hash   string
		method []string
	}
	tests := []struct {
		name       string
		args       args
		wantVerify bool
		wantErr    bool
	}{
		{
			name: "md5",
			args: args{
				file:   "ci.txt",
				hash:   "f621bae50c4c5099943aaaa3ef51c12b",
				method: []string{"md5"},
			},
			wantVerify: true,
			wantErr:    false,
		},
		{
			name: "sha1",
			args: args{
				file:   "ci.txt",
				hash:   "c560dd44beba55babb7e581de40861d8412d1cb6",
				method: []string{"sha1"},
			},
			wantVerify: true,
			wantErr:    false,
		},
		{
			name: "sha256",
			args: args{
				file:   "ci.txt",
				hash:   "9379c18887bd8cf8895c856bfbb1d4704a13595ff0a6cf4e7823770728cd6ee2",
				method: []string{"sha256"},
			},
			wantVerify: true,
			wantErr:    false,
		},
		{
			name: "sha512",
			args: args{
				file:   "ci.txt",
				hash:   "33500e835ae5009093d54ce1841a723e91a68664aa068a9d4493f1c168f67ab656792ce396de5d599e334e6d143fe8878f7371edea80e2f389511bdba7d095f4",
				method: []string{"sha512"},
			},
			wantVerify: true,
			wantErr:    false,
		},
		{
			name: "file not exist",
			args: args{
				file:   "not_exist.go",
				method: []string{"sha512"},
			},
			wantVerify: false,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVerify, err := FileVerify(tt.args.file, tt.args.hash, tt.args.method...)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashFileVerify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVerify != tt.wantVerify {
				t.Errorf("HashFileVerify() = %v, want %v", gotVerify, tt.wantVerify)
			}
		})
	}
}

func TestReader(t *testing.T) {
	type args struct {
		prepareReader func() (io.Reader, error)
		method        []string
	}
	tests := []struct {
		name     string
		args     args
		wantHash string
		wantErr  bool
	}{
		{
			name: "md5",
			args: args{
				prepareReader: func() (io.Reader, error) {
					return strings.NewReader("hello world"), nil
				},
				method: []string{"md5"},
			},
			wantHash: "5eb63bbbe01eeed093cb22bb8f5acdc3",
			wantErr:  false,
		},
		{
			name: "md5",
			args: args{
				prepareReader: func() (io.Reader, error) {
					return os.Open("ci.txt")
				},
				method: []string{"md5"},
			},
			wantHash: "f621bae50c4c5099943aaaa3ef51c12b",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := tt.args.prepareReader()
			if err != nil {
				t.Error(err)
			}
			gotHash, err := Reader(reader, tt.args.method...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHash != tt.wantHash {
				t.Errorf("Reader() = %v, want %v", gotHash, tt.wantHash)
			}
		})
	}
}
