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

package types

// GetSystemEndpointResponse ...
type GetSystemEndpointResponse struct {
	Endpoint string `json:"endpoint" example:"https://example.com:3000"`
}

// GetSystemVersionResponse ...
type GetSystemVersionResponse struct {
	Version   string `json:"version" example:"v1.0.0"`
	GitHash   string `json:"git_hash" example:"4225b69a"`
	BuildDate string `json:"build_date" example:"2023-10-16T11:25:45Z"`
}

// GetSystemConfigDaemon ...
type GetSystemConfigDaemon struct {
	Builder bool `json:"builder" example:"false"`
}

// GetSystemConfigResponse ...
type GetSystemConfigResponse struct {
	Daemon GetSystemConfigDaemon `json:"daemon"`
}
