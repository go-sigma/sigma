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

package authz

import (
	"go.uber.org/dig"

	"github.com/labstack/echo/v4/middleware"

	"github.com/go-sigma/sigma/pkg/types/enums"
)

// Config defines the config for CasbinAuth middleware
type Config struct {
	// Skipper defines a function to skip middleware
	Skipper middleware.Skipper
	// DigCon is the dig container
	DigCon *dig.Container
}

// AuthzConfig ...
type AuthzConfig struct {
	Skip    bool
	Sources []Source
}

// AuthzConfigSource ...
type Source struct {
	Scope    ScopeItem        `json:"scope"`
	Resource ResourceItem     `json:"resource"`
	Matched  bool             `json:"matched"`
	Effect   enums.AuthEffect `json:"pass"`
}

// ScopeItem ...
type ScopeItem struct {
	ScopeType  enums.AuthScope `json:"scope_type"`
	ScopeValue ScopeValue      `json:"scope_value"`
}

// ResourceItem ...
type ResourceItem struct {
	ResourceType  enums.AuthResource `json:"resource_type"`
	ResourceValue ResourceValue      `json:"resource_value"`
}

// ScopeValue ...
type ScopeValue struct {
	Position *enums.AuthPosition `json:"position"`
	Key      *string             `json:"key"`
	Value    string              `json:"value"`
}

// ResourceValue ...
type ResourceValue struct {
	Position *enums.AuthPosition `json:"position"`
	Key      *string             `json:"key"`
	Values   []string            `json:"values"`
}
