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

package api

// GetSystemEndpointResponse is the response for the public system endpoint.
type GetSystemEndpointResponse struct {
	Endpoint string `json:"endpoint" example:"https://example.com:3000"`
}

// GetSystemVersionResponse is the response for system version information.
type GetSystemVersionResponse struct {
	Version   string `json:"version" example:"v1.0.0"`
	GitHash   string `json:"git_hash" example:"4225b69a"`
	BuildDate string `json:"build_date" example:"2023-10-16T11:25:45Z"`
}

// GetSystemConfigDaemon describes daemon-related system configuration.
type GetSystemConfigDaemon struct {
	Builder bool `json:"builder" example:"false"`
}

// GetSystemConfigOAuth2 describes OAuth2-related system configuration.
type GetSystemConfigOAuth2 struct {
	GitHub bool `json:"github" example:"false"`
	GitLab bool `json:"gitlab" example:"false"`
}

// GetSystemConfigResponse is the response for public system configuration.
type GetSystemConfigResponse struct {
	Daemon    GetSystemConfigDaemon `json:"daemon"`
	Anonymous bool                  `json:"anonymous" example:"false"`
	OAuth2    GetSystemConfigOAuth2 `json:"oauth2"`
}
