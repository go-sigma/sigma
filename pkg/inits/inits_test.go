// Copyright 2023 XImager
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

package inits

import (
	"fmt"
	"testing"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name    string
		before  func()
		wantErr bool
	}{
		{
			name: "test-1",
			before: func() {
				inits["test-1"] = func() error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "test-2",
			before: func() {
				inits["test-2"] = func() error {
					return fmt.Errorf("test-2-error")
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inits = make(map[string]func() error)
			if tt.before != nil {
				tt.before()
			}
			if err := Initialize(); (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
