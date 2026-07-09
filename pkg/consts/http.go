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

package consts

// HTTP header names commonly used across the codebase.
//
// These mirror the canonical form defined by net/http
// (see http.CanonicalHeaderKey). Use these constants instead of
// raw string literals to keep header references centralized and
// typo-proof.
const (
	// HeaderContentType is the canonical "Content-Type" header.
	HeaderContentType = "Content-Type"
	// HeaderContentLength is the canonical "Content-Length" header.
	HeaderContentLength = "Content-Length"
	// HeaderContentRange is the canonical "Content-Range" header.
	HeaderContentRange = "Content-Range"
	// HeaderAccept is the canonical "Accept" header.
	HeaderAccept = "Accept"
	// HeaderAuthorization is the canonical "Authorization" header.
	HeaderAuthorization = "Authorization"
	// HeaderWWWAuthenticate is the canonical "WWW-Authenticate" header.
	HeaderWWWAuthenticate = "WWW-Authenticate"
	// HeaderLocation is the canonical "Location" header.
	HeaderLocation = "Location"
	// HeaderRange is the canonical "Range" header.
	HeaderRange = "Range"
	// HeaderETag is the canonical "ETag" header.
	HeaderETag = "ETag"
	// HeaderIfNoneMatch is the canonical "If-None-Match" header.
	HeaderIfNoneMatch = "If-None-Match"
	// HeaderUserAgent is the canonical "User-Agent" header.
	HeaderUserAgent = "User-Agent"
)
