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

package xerrors

import (
	"net/http"

	"github.com/labstack/echo/v4"
	dtspecv1 "github.com/opencontainers/distribution-spec/specs-go/v1"
)

// DSErrCode provides relevant information about a given error code.
type DSErrCode struct {
	// Value provides a unique, string key, often capitalized with
	// underscores, to identify the error code. This value is used as the
	// keyed value when serializing api errors.
	Value string

	// Message is a short, human readable description of the error condition
	// included in API responses.
	Message string

	// Description provides a complete account of the errors purpose, suitable
	// for use in documentation.
	Description string

	// HTTPStatusCode provides the http status code that is associated with
	// this error condition.
	HTTPStatusCode int
}

// NewDSError generates a distribution-spec error response
func NewDSError(c echo.Context, code DSErrCode) error {
	return c.JSON(code.HTTPStatusCode,
		dtspecv1.ErrorResponse{Errors: []dtspecv1.ErrorInfo{
			{
				Code:    code.Value,
				Message: code.Message,
				Detail:  code.Description,
			},
		}})
}

var (
	// DSErrCodeUnknown is a generic error that can be used as a last
	// resort if there is no situation-specific error message that can be used
	DSErrCodeUnknown = DSErrCode{
		Value:          "UNKNOWN",
		Message:        "unknown error",
		Description:    `Generic error returned when the error does not have an API classification.`,
		HTTPStatusCode: http.StatusInternalServerError,
	}

	// DSErrCodeUnsupported is returned when an operation is not supported.
	DSErrCodeUnsupported = DSErrCode{
		Value:          "UNSUPPORTED",
		Message:        "The operation is unsupported.",
		Description:    `The operation was unsupported due to a missing implementation or invalid set of parameters.`,
		HTTPStatusCode: http.StatusMethodNotAllowed,
	}

	// DSErrCodeUnauthorized is returned if a request requires
	// authentication.
	DSErrCodeUnauthorized = DSErrCode{
		Value:          "UNAUTHORIZED",
		Message:        "authentication required",
		Description:    `The access controller was unable to authenticate the client. Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate.`,
		HTTPStatusCode: http.StatusUnauthorized,
	}

	// DSErrCodeDenied is returned if a client does not have sufficient
	// permission to perform an action.
	DSErrCodeDenied = DSErrCode{
		Value:          "DENIED",
		Message:        "requested access to the resource is denied",
		Description:    `The access controller denied access for the operation on a resource.`,
		HTTPStatusCode: http.StatusForbidden,
	}

	// DSErrCodeUnavailable provides a common error to report unavailability
	// of a service or endpoint.
	DSErrCodeUnavailable = DSErrCode{
		Value:          "UNAVAILABLE",
		Message:        "service unavailable",
		Description:    "Returned when a service is not available",
		HTTPStatusCode: http.StatusServiceUnavailable,
	}

	// DSErrCodeTooManyRequests is returned if a client attempts too many
	// times to contact a service endpoint.
	DSErrCodeTooManyRequests = DSErrCode{
		Value:          "TOOMANYREQUESTS",
		Message:        "too many requests",
		Description:    `Returned when a client attempts to contact a service too many times`,
		HTTPStatusCode: http.StatusTooManyRequests,
	}

	// DSErrCodeDigestInvalid is returned when uploading a blob if the
	// provided digest does not match the blob contents.
	DSErrCodeDigestInvalid = DSErrCode{
		Value:          "DIGEST_INVALID",
		Message:        "provided digest did not match uploaded content",
		Description:    `When a blob is uploaded, the registry will check that the content matches the digest provided by the client. The error may include a detail structure with the key "digest", including the invalid digest string. This error may also be returned when a manifest includes an invalid layer digest.`,
		HTTPStatusCode: http.StatusBadRequest,
	}

	// DSErrCodeSizeInvalid is returned when uploading a blob if the provided
	DSErrCodeSizeInvalid = DSErrCode{
		Value:          "SIZE_INVALID",
		Message:        "provided length did not match content length",
		Description:    `When a layer is uploaded, the provided size will be checked against the uploaded content. If they do not match, this error will be returned.`,
		HTTPStatusCode: http.StatusBadRequest,
	}

	// DSErrCodeRangeInvalid is returned when uploading a blob if the provided
	// content range is invalid.
	DSErrCodeRangeInvalid = DSErrCode{
		Value:          "RANGE_INVALID",
		Message:        "invalid content range",
		Description:    `When a layer is uploaded, the provided range is checked against the uploaded chunk. This error is returned if the range is out of order.`,
		HTTPStatusCode: http.StatusRequestedRangeNotSatisfiable,
	}

	// DSErrCodeNameInvalid is returned when the name in the manifest does not
	// match the provided name.
	DSErrCodeNameInvalid = DSErrCode{
		Value:          "NAME_INVALID",
		Message:        "invalid repository name",
		Description:    `Invalid repository name encountered either during manifest validation or any API operation.`,
		HTTPStatusCode: http.StatusBadRequest,
	}

	// DSErrCodeTagInvalid is returned when the tag in the manifest does not
	// match the provided tag.
	DSErrCodeTagInvalid = DSErrCode{
		Value:          "TAG_INVALID",
		Message:        "manifest tag did not match URI",
		Description:    `During a manifest upload, if the tag in the manifest does not match the uri tag, this error will be returned.`,
		HTTPStatusCode: http.StatusBadRequest,
	}

	// DSErrCodeNameUnknown when the repository name is not known.
	DSErrCodeNameUnknown = DSErrCode{
		Value:          "NAME_UNKNOWN",
		Message:        "repository name not known to registry",
		Description:    `This is returned if the name used during an operation is unknown to the registry.`,
		HTTPStatusCode: http.StatusNotFound,
	}

	// DSErrCodeManifestUnknown returned when image manifest is unknown.
	DSErrCodeManifestUnknown = DSErrCode{
		Value:          "MANIFEST_UNKNOWN",
		Message:        "manifest unknown",
		Description:    `This error is returned when the manifest, identified by name and tag is unknown to the repository.`,
		HTTPStatusCode: http.StatusNotFound,
	}

	// DSErrCodeManifestInvalid returned when an image manifest is invalid,
	// typically during a PUT operation. This error encompasses all errors
	// encountered during manifest validation that aren't signature errors.
	DSErrCodeManifestInvalid = DSErrCode{
		Value:          "MANIFEST_INVALID",
		Message:        "manifest invalid",
		Description:    `During upload, manifests undergo several checks ensuring validity. If those checks fail, this error may be returned, unless a more specific error is included. The detail will contain information the failed validation.`,
		HTTPStatusCode: http.StatusBadRequest,
	}

	// DSErrCodeManifestUnverified is returned when the manifest fails
	// signature verification.
	DSErrCodeManifestUnverified = DSErrCode{
		Value:          "MANIFEST_UNVERIFIED",
		Message:        "manifest failed signature verification",
		Description:    `During manifest upload, if the manifest fails signature verification, this error will be returned.`,
		HTTPStatusCode: http.StatusBadRequest,
	}

	// DSErrCodeManifestBlobUnknown is returned when a manifest blob is
	// unknown to the registry.
	DSErrCodeManifestBlobUnknown = DSErrCode{
		Value:          "MANIFEST_BLOB_UNKNOWN",
		Message:        "blob unknown to registry",
		Description:    `This error may be returned when a manifest blob is unknown to the registry.`,
		HTTPStatusCode: http.StatusBadRequest,
	}

	// DSErrCodeBlobUnknown is returned when a blob is unknown to the
	// registry. This can happen when the manifest references a nonexistent
	// layer or the result is not found by a blob fetch.
	DSErrCodeBlobUnknown = DSErrCode{
		Value:          "BLOB_UNKNOWN",
		Message:        "blob unknown to registry",
		Description:    `This error may be returned when a blob is unknown to the registry in a specified repository. This can be returned with a standard get or if a manifest references an unknown layer during upload.`,
		HTTPStatusCode: http.StatusNotFound,
	}

	// DSErrCodeBlobUploadUnknown is returned when an upload is unknown.
	DSErrCodeBlobUploadUnknown = DSErrCode{
		Value:          "BLOB_UPLOAD_UNKNOWN",
		Message:        "blob upload unknown to registry",
		Description:    `If a blob upload has been cancelled or was never started, this error code may be returned.`,
		HTTPStatusCode: http.StatusNotFound,
	}

	// DSErrCodeBlobUploadInvalid is returned when an upload is invalid.
	DSErrCodeBlobUploadInvalid = DSErrCode{
		Value:          "BLOB_UPLOAD_INVALID",
		Message:        "blob upload invalid",
		Description:    `The blob upload encountered an error and can no longer proceed.`,
		HTTPStatusCode: http.StatusNotFound,
	}

	// DSErrCodeBlobUploadDigestMismatch is returned when an upload digest mismatch.
	DSErrCodeBlobUploadDigestMismatch = DSErrCode{
		Value:          "BLOB_UPLOAD_DIGEST_MISMATCH",
		Message:        "blob upload digest mismatch",
		Description:    `The blob upload digest mismatch.`,
		HTTPStatusCode: http.StatusNotFound,
	}

	// DSErrCodePaginationNumberInvalid is returned when the `n` parameter is
	// not an integer, or `n` is negative.
	DSErrCodePaginationNumberInvalid = DSErrCode{
		Value:          "PAGINATION_NUMBER_INVALID",
		Message:        "invalid number of results requested",
		Description:    `Returned when the "n" parameter (number of results to return) is not an integer, or "n" is negative.`,
		HTTPStatusCode: http.StatusBadRequest,
	}
)
