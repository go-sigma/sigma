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

package crypt

import (
	"testing"
)

func TestEncrypt(t *testing.T) {
	type args struct {
		key       string
		plaintext string
	}
	tests := []struct {
		name    string
		args    args
		want    func(*testing.T, string, string) string
		wantErr bool
	}{
		{
			name: "common",
			args: args{
				key:       "sigma",
				plaintext: "sigma",
			},
			want: func(t *testing.T, ciphertext, key string) string {
				plaintext, err := Decrypt(key, ciphertext)
				if err != nil {
					t.Errorf("decrypt failed: %v", err)
				}
				return plaintext
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encrypt(tt.args.key, tt.args.plaintext)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && tt.args.plaintext != tt.want(t, got, tt.args.key) {
				t.Errorf("Encrypt() = %v, want %v", got, tt.want(t, got, tt.args.key))
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	type args struct {
		key        string
		ciphertext string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "common",
			args: args{
				key:        "sigma",
				ciphertext: "uIjWyiYunVcb6aLRw5vaeIavFq1K",
			},
			want:    "sigma",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decrypt(tt.args.key, tt.args.ciphertext)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Decrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}
