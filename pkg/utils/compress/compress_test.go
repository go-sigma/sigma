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

package compress

import (
	"reflect"
	"testing"
)

func TestCompress(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "TestCompress-1",
			args: args{
				src: "test.txt",
			},
			want:    []byte{31, 139, 8, 0, 0, 0, 0, 0, 4, 255, 0, 15, 0, 240, 255, 104, 101, 108, 108, 111, 32, 115, 105, 103, 109, 97, 33, 33, 33, 10, 1, 0, 0, 255, 255, 38, 36, 146, 114, 15, 0, 0, 0},
			wantErr: false,
		},
		{
			name: "TestCompress-2",
			args: args{
				src: "no-exist.txt",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Compress(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Compress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecompress(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "TestDecompress-1",
			args: args{
				src: []byte{31, 139, 8, 0, 0, 0, 0, 0, 4, 255, 0, 15, 0, 240, 255, 104, 101, 108, 108, 111, 32, 115, 105, 103, 109, 97, 33, 33, 33, 10, 1, 0, 0, 255, 255, 38, 36, 146, 114, 15, 0, 0, 0},
			},
			want:    "hello sigma!!!\n",
			wantErr: false,
		},
		{
			name: "TestDecompress-1",
			args: args{
				src: []byte{11, 139, 8, 0, 0, 0, 0, 0, 4, 255, 0, 17, 0, 238, 255, 104, 101, 108, 108, 111, 32, 120, 105, 109, 97, 103, 101, 114, 33, 33, 33, 10, 1, 0, 0, 255, 255, 170, 130, 92, 143, 17, 0, 0, 0},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decompress(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decompress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Decompress() = %v, want %v", got, tt.want)
			}
		})
	}
}
