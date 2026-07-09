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

package bootstrap

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/config"
)

func TestInitialize(t *testing.T) {
	originalBootstrapTasks := bootstrapTasks
	t.Cleanup(func() {
		bootstrapTasks = originalBootstrapTasks
	})

	tests := []struct {
		name    string
		before  func()
		wantErr bool
	}{
		{
			name: "test-1",
			before: func() {
				bootstrapTasks = []bootstrapTask{
					{
						name: "test-1",
						run: func(*dig.Container) error {
							return nil
						},
					},
				}
			},
			wantErr: false,
		},
		{
			name: "test-2",
			before: func() {
				bootstrapTasks = []bootstrapTask{
					{
						name: "test-2",
						run: func(*dig.Container) error {
							return fmt.Errorf("test-2-error")
						},
					},
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.before != nil {
				tt.before()
			}
			digCon := dig.New()
			err := digCon.Provide(func() *config.Configuration { return &config.Configuration{} })
			assert.NoError(t, err)
			if err := Initialize(digCon); (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
