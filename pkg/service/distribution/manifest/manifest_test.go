// Copyright 2026 sigma
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

package manifest

import (
	"testing"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/manifest/manifestlist"
	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/opencontainers/go-digest"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api/enums"
)

func TestParseManifestRef(t *testing.T) {
	tag, dgst := parseManifestRef("latest")
	require.Equal(t, "latest", tag)
	require.Empty(t, dgst)

	expected := digest.FromString("manifest")
	tag, dgst = parseManifestRef(expected.String())
	require.Empty(t, tag)
	require.Equal(t, expected, dgst)
}

func TestManifestClassification(t *testing.T) {
	service := &service{}

	require.Equal(t, enums.ArtifactTypeImage, service.getArtifactType(
		distribution.Descriptor{MediaType: manifestlist.MediaTypeManifestList},
		fakeManifest{},
	))
	require.Equal(t, enums.ArtifactTypeImage, service.getArtifactType(
		distribution.Descriptor{},
		fakeManifest{references: []distribution.Descriptor{{MediaType: schema2.MediaTypeImageConfig}}},
	))
	require.Equal(t, enums.ArtifactTypeProvenance, service.getArtifactType(
		distribution.Descriptor{},
		fakeManifest{references: []distribution.Descriptor{{MediaType: inTotoMediaType}}},
	))
	require.Equal(t, enums.ArtifactTypeUnknown, service.getArtifactType(
		distribution.Descriptor{},
		fakeManifest{},
	))

	require.True(t, needScan(
		fakeManifest{references: []distribution.Descriptor{{MediaType: imgspecv1.MediaTypeImageConfig}}},
		distribution.Descriptor{},
	))
	require.False(t, needScan(
		fakeManifest{references: []distribution.Descriptor{{MediaType: cosignSimpleSigningMediaType}}},
		distribution.Descriptor{},
	))
}

type fakeManifest struct {
	references []distribution.Descriptor
}

func (f fakeManifest) References() []distribution.Descriptor {
	return f.references
}

func (fakeManifest) Payload() (string, []byte, error) {
	return "", nil, nil
}
