// Copyright 2024 sigma
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

package builder

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenContainerID(t *testing.T) {
	assert.Equal(t, fmt.Sprintf("%s%d-%d", ContainerPrefix, 1, 1), GenContainerID(1, 1))
}

func TestParseContainerID(t *testing.T) {
	type args struct {
		containerName string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		want1   int64
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				containerName: GenContainerID(1, 1),
			},
			want:    1,
			want1:   1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ParseContainerID(tt.args.containerName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseContainerID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseContainerID() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseContainerID() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
