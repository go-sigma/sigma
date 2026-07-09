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

import (
	"github.com/go-sigma/sigma/pkg/api/enums"
)

// GetUserListRequest is the request for listing users.
type GetUserListRequest struct {
	Pagination
	Sortable

	// Name filters users by username.
	Name         *string `json:"name" query:"name"`
	WithoutAdmin bool    `json:"without_admin" query:"without_admin"`
}

// PostUserRequest is the request for creating a user.
type PostUserRequest struct {
	Username       string         `json:"username" validate:"required,is_valid_username,min=2,max=20" example:"sigma"`
	Password       string         `json:"password" validate:"required,min=5,max=20,is_valid_password" example:"Admin@123"`
	Email          string         `json:"email" validate:"required,is_valid_email" example:"test@gmail.com"`
	NamespaceLimit *int64         `json:"namespace_limit,omitempty" validate:"omitempty,min=0" example:"10"`
	Role           enums.UserRole `json:"role" validate:"omitempty,is_valid_user_role" example:"Admin"`
}

// PutUserRequest is the request for updating a user.
type PutUserRequest struct {
	UserID string `param:"id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000" swaggerignore:"true"`

	Username       *string           `json:"username,omitempty" validate:"omitempty,is_valid_username,min=2,max=20" example:"sigma"`
	Password       *string           `json:"password,omitempty" validate:"omitempty,min=5,max=20,is_valid_password" example:"Admin@123"`
	Email          *string           `json:"email,omitempty" validate:"omitempty,is_valid_email" example:"test@gmail.com"`
	Status         *enums.UserStatus `json:"status,omitempty" validate:"omitempty,is_valid_user_status" example:"Active"`
	NamespaceLimit *int64            `json:"namespace_limit,omitempty" validate:"omitempty,min=0" example:"10"`
	Role           *enums.UserRole   `json:"role,omitempty" validate:"omitempty,is_valid_user_role" example:"Admin"`
}

// UserItem represents a user returned by the API.
type UserItem struct {
	ID             string           `json:"id"`
	Username       string           `json:"username"`
	Email          string           `json:"email"`
	Status         enums.UserStatus `json:"status"`
	LastLogin      string           `json:"last_login"`
	Role           enums.UserRole   `json:"role"`
	NamespaceLimit int64            `json:"namespace_limit"`
	NamespaceCount int64            `json:"namespace_count"`

	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// PostUserLoginRequest is the request for logging in.
type PostUserLoginRequest struct {
	Username string `json:"username" validate:"required,is_valid_username,min=2,max=20" example:"sigma"`
	Password string `json:"password" validate:"required,min=5,max=20,is_valid_password" example:"Admin@123"`
}

// PostUserLoginResponse is the response returned after a successful login.
type PostUserLoginResponse struct {
	RefreshToken string `json:"refresh_token"`
	Token        string `json:"token"`
	ID           string `json:"id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
}

// PostUserTokenResponse is the response returned after issuing an access token.
type PostUserTokenResponse struct {
	Token     string `json:"token" example:"sample-access-token"`
	ExpiresIn int    `json:"expires_in" example:"3600"`
	IssuedAt  string `json:"issued_at" example:"2023-07-16T17:51:51+08:00"`
}

// PostUserSignupRequest is the request for signing up.
type PostUserSignupRequest struct {
	Username string `json:"username" validate:"required,is_valid_username,min=2,max=20" example:"sigma"`
	Password string `json:"password" validate:"required,min=5,max=20,is_valid_password" example:"sigma2023X"`
	Email    string `json:"email" validate:"required,is_valid_email" example:"test@gmail.com"`
}

// PostUserSignupResponse is the response returned after a successful signup.
type PostUserSignupResponse struct {
	RefreshToken string `json:"refresh_token"`
	Token        string `json:"token"`
}

// GetUserSelfResponse is the response for the current user profile.
type GetUserSelfResponse struct {
	ID       string         `json:"id"`
	Email    string         `json:"email"`
	Username string         `json:"username"`
	Role     enums.UserRole `json:"role"`
}

// PostUserLogoutRequest is the request for logging out.
type PostUserLogoutRequest struct {
	Tokens []string `json:"tokens" validate:"required,min=1" example:"123,234"`
}

// PostUserRecoverPasswordRequest is the request for starting password recovery.
type PostUserRecoverPasswordRequest struct {
	Username string `json:"username" validate:"required,is_valid_username,min=2,max=20" example:"test"`
	Email    string `json:"email" validate:"required,is_valid_email" email:"test@email.com"`
}

// PostUserRecoverResetPasswordRequest is the request for resetting a recovered password.
type PostUserRecoverResetPasswordRequest struct {
	Code     string `json:"code" param:"code" validate:"required" example:"123456"`
	Password string `json:"password" validate:"required,min=5,max=20,is_valid_password" example:"sigma2023X"`
}

// PutUserSelfResetPasswordRequest is the request for resetting the current user's password.
type PutUserSelfResetPasswordRequest struct {
	Password string `json:"password" validate:"required,min=5,max=20,is_valid_password" example:"Admin@123"`
}

// PostUserResetPasswordPasswordRequest is the request for resetting another user's password.
type PostUserResetPasswordPasswordRequest struct {
	ID       string `json:"id" param:"id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Password string `json:"password" validate:"required,min=5,max=20,is_valid_password" example:"sigma2023X"`
}

// PutUserSelfRequest is the request for updating the current user's profile.
type PutUserSelfRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,is_valid_username,min=2,max=20" example:"sigma"`
	Email    *string `json:"email,omitempty" validate:"omitempty,is_valid_email" example:"test@mail.com"`
}

// ListCodeRepositoryProvidersResponse represents an enabled source code provider.
type ListCodeRepositoryProvidersResponse struct {
	Provider enums.Provider `json:"provider" example:"github"`
}

// GetCodeRepositoryResyncRequest is the request for resynchronizing source code repositories.
type GetCodeRepositoryResyncRequest struct {
	Provider enums.Provider `json:"provider" param:"provider" validate:"required,is_valid_provider"`
}

// GetCodeRepositoryUser3rdPartyRequest is the request for getting third-party source code account sync state.
type GetCodeRepositoryUser3rdPartyRequest struct {
	Provider enums.Provider `json:"provider" param:"provider" validate:"required,is_valid_provider"`
}

// GetCodeRepositoryUser3rdPartyResponse is the response for third-party source code account sync state.
type GetCodeRepositoryUser3rdPartyResponse struct {
	ID                    string                 `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	AccountID             string                 `json:"account_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CrLastUpdateTimestamp string                 `json:"cr_last_update_timestamp"`
	CrLastUpdateStatus    enums.TaskCommonStatus `json:"cr_last_update_status"`
	CrLastUpdateMessage   *string                `json:"cr_last_update_message"`

	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}
