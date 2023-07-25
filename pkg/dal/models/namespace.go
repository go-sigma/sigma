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

package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/go-sigma/sigma/pkg/types/enums"
)

// Namespace represents a namespace
type Namespace struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	Name            string `gorm:"uniqueIndex"`
	Description     *string
	Visibility      enums.Visibility `gorm:"default:public"`
	TagLimit        int64            `gorm:"default:0"`
	TagCount        int64            `gorm:"default:0"`
	RepositoryLimit int64            `gorm:"default:0"`
	RepositoryCount int64            `gorm:"default:0"`
	SizeLimit       int64            `gorm:"default:0"`
	Size            int64            `gorm:"default:0"`
}

var policyStatement1 = "INSERT INTO `casbin_rules` (`ptype`, `v0`, `v1`, `v2`, `v3`, `v4`) VALUES ('p', '^_^Namespace^_^_admin', '/namespaces/^_^Namespace^_^', '*', 'allow');"

var policyStatements = []string{}

func init() {
	policyStatements = append(policyStatements, policyStatement1)
}

func policyStatement(namespaceName string) string {
	var result string
	for _, p := range policyStatements {
		result += strings.ReplaceAll(p, "^_^Namespace^_^", namespaceName)
	}
	return result
}

// BeforeCreate ...
func (n *Namespace) BeforeCreate(tx *gorm.DB) error {
	if n == nil || n.ID == 0 {
		return nil
	}
	return tx.Exec(policyStatement(n.Name)).Error
}
