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

package main

import (
	"gorm.io/gen"

	"github.com/go-sigma/sigma/pkg/dal/models"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:       "pkg/dal/query",
		Mode:          gen.WithDefaultQuery,
		FieldSignable: true,
	})

	g.ApplyBasic(
		models.User{},
		models.User3rdParty{},
		models.UserRecoverCode{},
		models.CodeRepository{},
		models.CodeRepositoryBranch{},
		models.CodeRepositoryOwner{},
		models.CodeRepositoryCloneCredential{},
		models.Audit{},
		models.Namespace{},
		models.Repository{},
		models.Artifact{},
		models.ArtifactSbom{},
		models.ArtifactVulnerability{},
		models.Tag{},
		models.Blob{},
		models.BlobUpload{},
		models.CasbinRule{},
		models.Webhook{},
		models.WebhookLog{},
		models.Builder{},
		models.BuilderRunner{},
		models.WorkQueue{},
		models.Setting{},
		models.DaemonGcTagRule{},
		models.DaemonGcTagRunner{},
		models.DaemonGcTagRecord{},
		models.DaemonGcRepositoryRule{},
		models.DaemonGcRepositoryRunner{},
		models.DaemonGcRepositoryRecord{},
		models.DaemonGcArtifactRule{},
		models.DaemonGcArtifactRunner{},
		models.DaemonGcArtifactRecord{},
		models.DaemonGcBlobRule{},
		models.DaemonGcBlobRunner{},
		models.DaemonGcBlobRecord{},
		models.NamespaceMember{},
	)

	g.ApplyInterface(func(models.ArtifactSizeByNamespaceOrRepository) {}, models.Artifact{})
	g.ApplyInterface(func(models.ArtifactAssociated) {}, models.Artifact{})
	g.ApplyInterface(func(models.BlobAssociateWithArtifact) {}, models.Blob{})

	g.Execute()
}
