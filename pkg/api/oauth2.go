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

import "github.com/go-sigma/sigma/pkg/api/enums"

// Oauth2LoginRequest is the request for starting an OAuth2 login flow.
type Oauth2LoginRequest struct {
	Provider enums.Provider `json:"provider" param:"provider" validate:"required,is_valid_provider"`
}

// Oauth2CallbackRequest is the request for handling an OAuth2 callback.
type Oauth2CallbackRequest struct {
	Provider enums.Provider `json:"provider" param:"provider" validate:"required,is_valid_provider" example:"gitlab"`
	Code     string         `json:"code" query:"code" validate:"required" example:"123456"`
	Endpoint string         `json:"endpoint" query:"endpoint" example:"http://localhost:5173"`
}

// Oauth2RedirectCallbackRequest is the request for handling an OAuth2 redirect callback.
type Oauth2RedirectCallbackRequest struct {
	Code string `json:"code" query:"code"`
}

// Oauth2UserInfo represents user information returned by an OAuth2 provider.
type Oauth2UserInfo struct {
	Provider     enums.Provider `json:"provider"`
	ID           string         `json:"id"`
	Username     string         `json:"username"`
	Email        string         `json:"email"`
	Token        string         `json:"token"`
	RefreshToken string         `json:"refresh_token"`
}

// Oauth2CallbackResponse is the response returned after a successful OAuth2 callback.
type Oauth2CallbackResponse struct {
	RefreshToken string `json:"refresh_token" example:"sample-refresh-token"`
	Token        string `json:"token" example:"sample-access-token"`
	ID           string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email        string `json:"email" example:"test@email.com"`
	Username     string `json:"username" example:"sigma"`
}

// Oauth2ClientIDRequest is the request for getting an OAuth2 client ID.
type Oauth2ClientIDRequest struct {
	Provider enums.Provider `json:"provider" param:"provider" validate:"required,is_valid_provider"`
}

// Oauth2ClientIDResponse is the response containing an OAuth2 client ID.
type Oauth2ClientIDResponse struct {
	ClientID string `json:"client_id" example:"1234567890"`
}
