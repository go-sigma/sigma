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

package daemon

import "testing"

func TestMustParseDaemonPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustParseDaemon did not panic")
		}
	}()
	MustParseDaemon("InvalidDaemon")
}

func TestMustParseDaemon(t *testing.T) {
	MustParseDaemon("Vulnerability")
	MustParseDaemon("Sbom")
	MustParseDaemon("ProxyArtifact")
	MustParseDaemon("ProxyTag")
}

func TestIsValid(t *testing.T) {
	if !DaemonVulnerability.IsValid() {
		t.Errorf("DaemonVulnerability is not valid")
	}
	if !DaemonSbom.IsValid() {
		t.Errorf("DaemonSbom is not valid")
	}
	if !DaemonProxyArtifact.IsValid() {
		t.Errorf("DaemonProxyArtifact is not valid")
	}
	if !DaemonProxyTag.IsValid() {
		t.Errorf("DaemonProxyTag is not valid")
	}
}

func TestString(t *testing.T) {
	if DaemonVulnerability.String() != "Vulnerability" {
		t.Errorf("DaemonVulnerability does not match")
	}
	if DaemonSbom.String() != "Sbom" {
		t.Errorf("DaemonSbom does not match")
	}
	if DaemonProxyArtifact.String() != "ProxyArtifact" {
		t.Errorf("DaemonProxyArtifact does not match")
	}
	if DaemonProxyTag.String() != "ProxyTag" {
		t.Errorf("DaemonProxyTag does not match")
	}
}
