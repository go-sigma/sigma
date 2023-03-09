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

package dal

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/consts"
)

// AuthEnforcer is the global casbin enforcer
var AuthEnforcer *casbin.Enforcer

func setAuthModel(db *gorm.DB) error {
	authModel, err := model.NewModelFromString(consts.AuthModel)
	if err != nil {
		return err
	}
	adapter, err := gormadapter.NewAdapterByDBUseTableName(db, "", "casbin_rules")
	if err != nil {
		return err
	}
	AuthEnforcer, err = casbin.NewEnforcer(authModel, adapter)
	if err != nil {
		return err
	}
	return nil
}
