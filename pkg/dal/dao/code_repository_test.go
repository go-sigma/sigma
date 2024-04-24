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

package dao_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/query"
)

func TestCodeRepositoryServiceFactory(t *testing.T) {
	f := dao.NewCodeRepositoryServiceFactory()
	assert.NotNil(t, f.New())
	assert.NotNil(t, f.New(query.Q))
}
