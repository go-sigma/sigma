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

package xerrors

import (
	"fmt"
	"math"
	"math/big"
	"net/http"

	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v4"
	dtspecv1 "github.com/opencontainers/distribution-spec/specs-go/v1"
)

// NewDSError generates a distribution-spec error response
func NewDSError(c echo.Context, code ErrCode) error {
	return c.JSON(code.HTTPStatusCode,
		dtspecv1.ErrorResponse{Errors: []dtspecv1.ErrorInfo{
			{
				Code:    code.Code,
				Message: code.Title,
				Detail:  code.Description,
			},
		}})
}

var (
	// DSErrCodeUnknown is a generic error that can be used as a last
	// resort if there is no situation-specific error message that can be used
	DSErrCodeUnknown = ErrCode{
		Code:           "UNKNOWN",
		Title:          "unknown error",
		Description:    `Generic error returned when the error does not have an API classification.`,
		HTTPStatusCode: http.StatusInternalServerError,
	}

	// DSErrCodeUnsupported is returned when an operation is not supported.
	DSErrCodeUnsupported = ErrCode{
		Code:           "UNSUPPORTED",
		Title:          "The operation is unsupported.",
		Description:    `The operation was unsupported due to a missing implementation or invalid set of parameters.`,
		HTTPStatusCode: http.StatusMethodNotAllowed,
	}

	// DSErrCodeUnauthorized is returned if a request requires
	// authentication.
	DSErrCodeUnauthorized = ErrCode{
		Code:           "UNAUTHORIZED",
		Title:          "authentication required",
		Description:    `The access controller was unable to authenticate the client. Often this will be accompanied by a Www-Authenticate HTTP response header indicating how to authenticate.`,
		HTTPStatusCode: http.StatusUnauthorized,
	}

	// DSErrCodeDenied is returned if a client does not have sufficient
	// permission to perform an action.
	DSErrCodeDenied = ErrCode{
		Code:           "DENIED",
		Title:          "requested access to the resource is denied",
		Description:    `The access controller denied access for the operation on a resource.`,
		HTTPStatusCode: http.StatusForbidden,
	}

	DSErrCodeResourceQuotaExceed = ErrCode{
		Code:           "DENIED",
		Title:          "requested access to the resource quota is exceed",
		Description:    `The access controller denied access for the operation on a resource.`,
		HTTPStatusCode: http.StatusForbidden,
	}

	// DSErrCodeUnavailable provides a common error to report unavailability
	// of a service or endpoint.
	DSErrCodeUnavailable = ErrCode{
		Code:           "UNAVAILABLE",
		Title:          "service unavailable",
		Description:    "Returned when a service is not available",
		HTTPStatusCode: http.StatusServiceUnavailable,
	}

	// DSErrCodeTooManyRequests is returned if a client attempts too many
	// times to contact a service endpoint.
	DSErrCodeTooManyRequests = ErrCode{
		Code:           "TOOMANYREQUESTS",
		Title:          "too many requests",
		Description:    `Returned when a client attempts to contact a service too many times`,
		HTTPStatusCode: http.StatusTooManyRequests,
	}

	// DSErrCodeDigestInvalid is returned when uploading a blob if the
	// provided digest does not match the blob contents.
	DSErrCodeDigestInvalid = ErrCode{
		Code:           "DIGEST_INVALID",
		Title:          "provided digest did not match uploaded content",
		Description:    `When a blob is uploaded, the registry will check that the content matches the digest provided by the client. The error may include a detail structure with the key "digest", including the invalid digest string. This error may also be returned when a manifest includes an invalid layer digest.`,
		HTTPStatusCode: http.StatusBadRequest,
	}

	// DSErrCodeSizeInvalid is returned when uploading a blob if the provided
	DSErrCodeSizeInvalid = ErrCode{
		Code:           "SIZE_INVALID",
		Title:          "provided length did not match content length",
		Description:    `When a layer is uploaded, the provided size will be checked against the uploaded content. If they do not match, this error will be returned.`,
		HTTPStatusCode: http.StatusBadRequest,
	}

	// DSErrCodeRangeInvalid is returned when uploading a blob if the provided
	// content range is invalid.
	DSErrCodeRangeInvalid = ErrCode{
		Code:           "RANGE_INVALID",
		Title:          "invalid content range",
		Description:    `When a layer is uploaded, the provided range is checked against the uploaded chunk. This error is returned if the range is out of order.`,
		HTTPStatusCode: http.StatusRequestedRangeNotSatisfiable,
	}

	// DSErrCodeNameInvalid is returned when the name in the manifest does not
	// match the provided name.
	DSErrCodeNameInvalid = ErrCode{
		Code:           "NAME_INVALID",
		Title:          "invalid repository name",
		Description:    `Invalid repository name encountered either during manifest validation or any API operation.`,
		HTTPStatusCode: http.StatusBadRequest,
	}

	// DSErrCodeTagInvalid is returned when the tag in the manifest does not
	// match the provided tag.
	DSErrCodeTagInvalid = ErrCode{
		Code:           "TAG_INVALID",
		Title:          "manifest tag did not match URI",
		Description:    `During a manifest upload, if the tag in the manifest does not match the uri tag, this error will be returned.`,
		HTTPStatusCode: http.StatusBadRequest,
	}

	// DSErrCodeNameUnknown when the repository name is not known.
	DSErrCodeNameUnknown = ErrCode{
		Code:           "NAME_UNKNOWN",
		Title:          "repository name not known to registry",
		Description:    `This is returned if the name used during an operation is unknown to the registry.`,
		HTTPStatusCode: http.StatusNotFound,
	}

	// DSErrCodeManifestUnknown returned when image manifest is unknown.
	DSErrCodeManifestUnknown = ErrCode{
		Code:           "MANIFEST_UNKNOWN",
		Title:          "manifest unknown",
		Description:    `This error is returned when the manifest, identified by name and tag is unknown to the repository.`,
		HTTPStatusCode: http.StatusNotFound,
	}

	// DSErrCodeManifestInvalid returned when an image manifest is invalid,
	// typically during a PUT operation. This error encompasses all errors
	// encountered during manifest validation that aren't signature errors.
	DSErrCodeManifestInvalid = ErrCode{
		Code:           "MANIFEST_INVALID",
		Title:          "manifest invalid",
		Description:    `During upload, manifests undergo several checks ensuring validity. If those checks fail, this error may be returned, unless a more specific error is included. The detail will contain information the failed validation.`,
		HTTPStatusCode: http.StatusBadRequest,
	}

	// DSErrCodeManifestUnverified is returned when the manifest fails
	// signature verification.
	DSErrCodeManifestUnverified = ErrCode{
		Code:           "MANIFEST_UNVERIFIED",
		Title:          "manifest failed signature verification",
		Description:    `During manifest upload, if the manifest fails signature verification, this error will be returned.`,
		HTTPStatusCode: http.StatusBadRequest,
	}

	// DSErrCodeManifestBlobUnknown is returned when a manifest blob is
	// unknown to the registry.
	DSErrCodeManifestBlobUnknown = ErrCode{
		Code:           "MANIFEST_BLOB_UNKNOWN",
		Title:          "blob unknown to registry",
		Description:    `This error may be returned when a manifest blob is unknown to the registry.`,
		HTTPStatusCode: http.StatusBadRequest,
	}

	// DSErrCodeManifestWithNamespace is returned when a manifest name is
	// invalid because it must contain a valid namespace.
	DSErrCodeManifestWithNamespace = ErrCode{
		Code:           "MANIFEST_WITH_NAMESPACE",
		Title:          "manifest name must contain a namespace",
		Description:    `This error may be returned when a manifest name is invalid because it is not contain a valid namespace, the image name must be like 'test.com/library/nginx:latest' or 'test.com/public/busybox:latest', but 'test.com/nginx:latest' is not allowed.`,
		HTTPStatusCode: http.StatusBadRequest,
	}

	// DSErrCodeBlobUnknown is returned when a blob is unknown to the
	// registry. This can happen when the manifest references a nonexistent
	// layer or the result is not found by a blob fetch.
	DSErrCodeBlobUnknown = ErrCode{
		Code:           "BLOB_UNKNOWN",
		Title:          "blob unknown to registry",
		Description:    `This error may be returned when a blob is unknown to the registry in a specified repository. This can be returned with a standard get or if a manifest references an unknown layer during upload.`,
		HTTPStatusCode: http.StatusNotFound,
	}

	// DSErrCodeBlobAssociated is returned when a blob upload is attempted
	// for a blob that is already referenced by another manifest.
	DSErrCodeBlobAssociated = ErrCode{
		Code:           "BLOB_ASSOCIATED",
		Title:          "blob associated with multiple manifests",
		Description:    `This error is returned when a blob is uploaded that is already referenced by another manifest in the repository.`,
		HTTPStatusCode: http.StatusBadRequest,
	}

	// DSErrCodeBlobUploadUnknown is returned when an upload is unknown.
	DSErrCodeBlobUploadUnknown = ErrCode{
		Code:           "BLOB_UPLOAD_UNKNOWN",
		Title:          "blob upload unknown to registry",
		Description:    `If a blob upload has been cancelled or was never started, this error code may be returned.`,
		HTTPStatusCode: http.StatusNotFound,
	}

	// DSErrCodeBlobUploadInvalid is returned when an upload is invalid.
	DSErrCodeBlobUploadInvalid = ErrCode{
		Code:           "BLOB_UPLOAD_INVALID",
		Title:          "blob upload invalid",
		Description:    `The blob upload encountered an error and can no longer proceed.`,
		HTTPStatusCode: http.StatusNotFound,
	}

	// DSErrCodeBlobUploadDigestMismatch is returned when an upload digest mismatch.
	DSErrCodeBlobUploadDigestMismatch = ErrCode{
		Code:           "BLOB_UPLOAD_DIGEST_MISMATCH",
		Title:          "blob upload digest mismatch",
		Description:    `The blob upload digest mismatch.`,
		HTTPStatusCode: http.StatusNotFound,
	}

	// DSErrCodePaginationNumberInvalid is returned when the `n` parameter is
	// not an integer, or `n` is negative.
	DSErrCodePaginationNumberInvalid = ErrCode{
		Code:           "PAGINATION_NUMBER_INVALID",
		Title:          "invalid number of results requested",
		Description:    `Returned when the "n" parameter (number of results to return) is not an integer, or "n" is negative.`,
		HTTPStatusCode: http.StatusBadRequest,
	}
)

// GenDSErrCodeResourceSizeQuotaExceedNamespace ...
func GenDSErrCodeResourceSizeQuotaExceedNamespace(name string, current, limit, increase int64) ErrCode {
	c := ErrCode{
		Code: "DENIED",
		Title: fmt.Sprintf("requested access to the size quota is exceed, namespace(%s) size quota %s/%s(%s%%), increasing size is %s",
			name,
			humanize.BigIBytes(big.NewInt(current)),
			humanize.BigIBytes(big.NewInt(limit)), humanize.Ftoa(toFixed(float64(current)/float64(limit)*100, 1)),
			humanize.BigIBytes(big.NewInt(increase))),
		Description:    `The access controller denied access for the operation on a resource.`,
		HTTPStatusCode: http.StatusForbidden,
	}
	return c
}

// GenDSErrCodeResourceSizeQuotaExceedRepository ...
func GenDSErrCodeResourceSizeQuotaExceedRepository(name string, current, limit, increase int64) ErrCode {
	c := ErrCode{
		Code: "DENIED",
		Title: fmt.Sprintf("requested access to the size quota is exceed, repository(%s) size quota %s/%s(%s%%), increasing size is %s",
			name,
			humanize.BigIBytes(big.NewInt(current)),
			humanize.BigIBytes(big.NewInt(limit)), humanize.Ftoa(toFixed(float64(current)/float64(limit)*100, 1)),
			humanize.BigIBytes(big.NewInt(increase))),
		Description:    `The access controller denied access for the operation on a resource.`,
		HTTPStatusCode: http.StatusForbidden,
	}
	return c
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

// GenDSErrCodeResourceCountQuotaExceedRepository ...
func GenDSErrCodeResourceCountQuotaExceedRepository(name string, limit int64) ErrCode {
	c := ErrCode{
		Code:           "DENIED",
		Title:          fmt.Sprintf("requested access to the resource count quota is exceed, repository(%s) tag count quota is %d", name, limit),
		Description:    `The access controller denied access for the operation on a resource.`,
		HTTPStatusCode: http.StatusForbidden,
	}
	return c
}

// GenDSErrCodeResourceCountQuotaExceedNamespaceRepository ...
func GenDSErrCodeResourceCountQuotaExceedNamespaceRepository(name string, limit int64) ErrCode {
	c := ErrCode{
		Code:           "DENIED",
		Title:          fmt.Sprintf("requested access to the resource count quota is exceed, namespace(%s) repository count quota is %d", name, limit),
		Description:    `The access controller denied access for the operation on a resource.`,
		HTTPStatusCode: http.StatusForbidden,
	}
	return c
}

// GenDSErrCodeResourceCountQuotaExceedNamespaceTag ...
func GenDSErrCodeResourceCountQuotaExceedNamespaceTag(name string, limit int64) ErrCode {
	c := ErrCode{
		Code:           "DENIED",
		Title:          fmt.Sprintf("requested access to the resource count quota is exceed, namespace(%s) tag count quota is %d", name, limit),
		Description:    `The access controller denied access for the operation on a resource.`,
		HTTPStatusCode: http.StatusForbidden,
	}
	return c
}

// GenDSErrCodeResourceNotFound ...
func GenDSErrCodeResourceNotFound(err error) ErrCode {
	c := ErrCode{
		Code:           "RESOURCE_NOT_FOUND",
		Title:          fmt.Sprintf("%v", err),
		Description:    `The request resource is not found.`,
		HTTPStatusCode: http.StatusNotFound,
	}
	return c
}
