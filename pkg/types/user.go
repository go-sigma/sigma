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

package types

// PostUserLoginRequest ...
type PostUserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// PostUserLoginResponse ...
type PostUserLoginResponse struct {
	RefreshToken string `json:"refresh_token"`
	Token        string `json:"token"`
}

// PostUserTokenRequest ...
type PostUserTokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	IssuedAt  string `json:"issued_at"`
}

// PostUserSignupRequest ...
type PostUserSignupRequest struct {
	Username string `json:"username" validate:"required,alphanum,min=2,max=20"`
	Password string `json:"password" validate:"required,min=6,max=20"`
	Email    string `json:"email" validate:"required,email"`
}

// PostUserSignupResponse ...
type PostUserSignupResponse struct {
	RefreshToken string `json:"refresh_token"`
	Token        string `json:"token"`
}
