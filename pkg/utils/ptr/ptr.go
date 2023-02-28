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

// Package ptr ...
package ptr

// Of returns pointer to value.
func Of[T any](v T) *T {
	return &v
}

// To returns the value of the pointer passed in or the default value if the pointer is nil.
func To[T any](v *T) T {
	var zero T
	if v == nil {
		return zero
	}
	return *v
}

// ToDef returns the value of the int pointer passed in or default value if the pointer is nil.
func ToDef[T any](v *T, def T) T {
	if v == nil {
		return def
	}
	return *v
}
