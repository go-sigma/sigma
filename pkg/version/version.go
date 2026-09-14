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

package version

var (
	// Version is the git describe tag of the build, injected at link time via -X github.com/go-sigma/sigma/pkg/version.Version.
	Version = ""
	// GitHash is the short git commit hash of the build, injected at link time via -X github.com/go-sigma/sigma/pkg/version.GitHash.
	GitHash = ""
	// BuildDate is the UTC build timestamp, injected at link time via -X github.com/go-sigma/sigma/pkg/version.BuildDate.
	BuildDate = ""
)
