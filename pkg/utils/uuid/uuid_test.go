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

package uuid

import "testing"

func TestNewV7StringReturnsOrderedUUIDV7(t *testing.T) {
	previous := NewV7String()
	parsed, err := Parse(previous)
	if err != nil {
		t.Fatalf("parse uuid: %v", err)
	}
	if parsed.Version() != 7 {
		t.Fatalf("uuid version = %d, want 7", parsed.Version())
	}

	for range 1000 {
		current := NewV7String()
		parsed, err := Parse(current)
		if err != nil {
			t.Fatalf("parse uuid: %v", err)
		}
		if parsed.Version() != 7 {
			t.Fatalf("uuid version = %d, want 7", parsed.Version())
		}
		if current <= previous {
			t.Fatalf("uuid order is not increasing: current=%s previous=%s", current, previous)
		}
		previous = current
	}
}
