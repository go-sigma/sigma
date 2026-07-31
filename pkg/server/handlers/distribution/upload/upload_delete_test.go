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

package upload

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	svcupload "github.com/go-sigma/sigma/pkg/service/distribution/upload"
)

func TestDeleteUpload(t *testing.T) {
	tests := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "deletes upload", wantStatus: http.StatusNoContent},
		{name: "maps service failure", serviceErr: errors.New("delete failed"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := svcupload.NewMockDistributionUploadService(ctrl)
			svc.EXPECT().DeleteUpload(gomock.Any(), "upload-1").Return(tt.serviceErr)
			recorder, c := newUploadContext(t, http.MethodDelete, "/v2/library/alpine/blobs/uploads/upload-1", nil)

			(&handler{UploadSvc: svc}).DeleteUpload(c)
			c.Writer.WriteHeaderNow()

			require.Equal(t, tt.wantStatus, recorder.Code)
		})
	}
}
