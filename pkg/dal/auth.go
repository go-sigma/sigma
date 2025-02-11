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

package dal

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
)

// AuthEnforcer is the global casbin enforcer
var AuthEnforcer *casbin.SyncedEnforcer

func setAuthModel(db *gorm.DB) error {
	authModel, err := model.NewModelFromString(consts.AuthModel)
	if err != nil {
		return err
	}
	gormadapter.TurnOffAutoMigrate(db)
	adapter, err := gormadapter.NewAdapterByDBUseTableName(db, "", "casbin_rules")
	if err != nil {
		return err
	}
	AuthEnforcer, err = casbin.NewSyncedEnforcer(authModel, adapter)
	if err != nil {
		return err
	}
	AuthEnforcer.StartAutoLoadPolicy(time.Minute * 10)
	AuthEnforcer.EnableAutoSave(true)
	AuthEnforcer.AddFunction("urlMatch", UrlMatchFunc)
	return nil
}

const (
	delimiter = "$"
)

// UrlMatchFunc ...
// DS${repository}$tags
// DS$catalog
// DS${repository}$(blobs)|(manifest)|(blob_uploads)${reference}
// DS${repository}$blob_uploads
// API${repository}${url}
func UrlMatchFunc(args ...any) (any, error) {
	request := args[0].(string)
	policy := args[1].(string)

	uRequest, err := url.Parse(request)
	if err != nil {
		return false, err
	}

	if strings.HasPrefix(request, "/v2/") {
		request = uRequest.Path
		if request == "/v2/" && policy == fmt.Sprintf("DS%sv2", delimiter) { // nolint: gocritic
			return true, nil
		} else if strings.HasSuffix(request, "/_catalog") {
			return policy == fmt.Sprintf("DS%scatalog", delimiter), nil
		} else if strings.HasSuffix(request, "tags/list") {
			repository := strings.TrimPrefix(strings.TrimSuffix(request, "/tags/list"), "/v2/")
			return pathPattern(fmt.Sprintf("DS$%s$tags", repository), policy)
		} else if strings.HasSuffix(request, "/blobs/uploads/") {
			repository := strings.TrimPrefix(strings.TrimSuffix(request, "/blobs/uploads/"), "/v2/")
			return pathPattern(fmt.Sprintf("DS$%s$blob_uploads", repository), policy)
		} else {
			rRequest := request[:strings.LastIndex(request, "/")]
			ref := strings.TrimPrefix(request[strings.LastIndex(request, "/"):], "/")
			if strings.HasSuffix(rRequest, "/manifests") { // nolint: gocritic
				repository := strings.TrimPrefix(strings.TrimSuffix(rRequest, "/manifests"), "/v2/")
				return pathPattern(fmt.Sprintf("DS$%s$manifests$%s", repository, ref), policy)
			} else if strings.HasSuffix(rRequest, "/blobs") {
				repository := strings.TrimPrefix(strings.TrimSuffix(rRequest, "/blobs"), "/v2/")
				return pathPattern(fmt.Sprintf("DS$%s$blobs$%s", repository, ref), policy)
			} else if strings.HasSuffix(rRequest, "/blobs/uploads") {
				repository := strings.TrimPrefix(strings.TrimSuffix(rRequest, "/blobs/uploads"), "/v2/")
				return pathPattern(fmt.Sprintf("DS$%s$blob_uploads$%s", repository, ref), policy)
			}
		}
	} else if strings.HasPrefix(request, "/api/") {
		repository := uRequest.Query().Get("repository")
		if repository != "" {
			return pathPattern(fmt.Sprintf("API$%s$%s", repository, strings.TrimPrefix(request, "/api/")), policy)
		} else {
			return pathPattern(fmt.Sprintf("API$%s", strings.TrimPrefix(request, "/api/")), policy)
		}
	}
	return false, nil
}

func pathPattern(request, policy string) (bool, error) {
	return filepath.Match(strings.Join(strings.Split(policy, "$"), "/"), strings.Join(strings.Split(request, "$"), "/"))
}
