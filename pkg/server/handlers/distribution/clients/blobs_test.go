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

package clients

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/hash"
)

func TestGetBlob(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	body := "test-busybox"
	hashStr, err := hash.String(body)
	assert.NoError(t, err)
	dgest, err := digest.Parse("sha256:" + hashStr)
	assert.NoError(t, err)
	mux.HandleFunc("/v2/library/busybox/blobs/"+dgest.String(), func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add(consts.ContentDigest, dgest.String())
		w.Header().Add(echo.HeaderContentType, "application/vnd.oci.image.layer.v1.tar+gzip")
		w.Write([]byte(body)) // nolint: errcheck
	})
	s := httptest.NewServer(mux)

	f := NewClientsFactory()
	cli, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  s.URL,
			TlsVerify: true,
		}})
	assert.NoError(t, err)

	descriptor, reader, err := cli.GetBlob(context.Background(), "library/busybox", dgest)
	assert.NoError(t, err)
	bodyBytes, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, body, string(bodyBytes))
	assert.Equal(t, descriptor.Digest.String(), dgest.String())
	assert.Equal(t, descriptor.MediaType, "application/vnd.oci.image.layer.v1.tar+gzip")
	assert.Equal(t, descriptor.Size, int64(len(body)))
}

func TestGetBlob1(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	body := "test-busybox"
	hashStr, err := hash.String(body)
	assert.NoError(t, err)
	dgest, err := digest.Parse("sha256:" + hashStr)
	assert.NoError(t, err)
	mux.HandleFunc("/v2/library/busybox/blobs/"+dgest.String(), func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add(consts.ContentDigest, dgest.String())
		w.Header().Add(echo.HeaderContentType, "application/vnd.oci.image.layer.v1.tar+gzip")
		w.Header().Add(echo.HeaderContentLength, "aaa")
		w.Write([]byte(body)) // nolint: errcheck
	})
	s := httptest.NewServer(mux)

	f := NewClientsFactory()
	cli, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  s.URL,
			TlsVerify: true,
		}})
	assert.NoError(t, err)

	_, _, err = cli.GetBlob(context.Background(), "library/busybox", dgest)
	assert.Error(t, err)
}

func TestGetBlob2(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	body := "test-busybox"
	hashStr, err := hash.String(body)
	assert.NoError(t, err)
	dgest, err := digest.Parse("sha256:" + hashStr)
	assert.NoError(t, err)
	mux.HandleFunc("/v2/library/busybox/blobs/"+dgest.String(), func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add(consts.ContentDigest, "m"+dgest.String())
		w.Header().Add(echo.HeaderContentType, "application/vnd.oci.image.layer.v1.tar+gzip")
		w.Write([]byte(body)) // nolint: errcheck
	})
	s := httptest.NewServer(mux)

	f := NewClientsFactory()
	cli, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  s.URL,
			TlsVerify: true,
		}})
	assert.NoError(t, err)

	_, _, err = cli.GetBlob(context.Background(), "library/busybox", dgest)
	assert.Error(t, err)
}

func TestGetBlob3(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	body := "test-busybox"
	hashStr, err := hash.String(body)
	assert.NoError(t, err)
	dgest, err := digest.Parse("sha256:" + hashStr)
	assert.NoError(t, err)
	mux.HandleFunc("/v2/library/busybox/blobs/"+dgest.String(), func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add(consts.ContentDigest, "m"+dgest.String())
		w.Header().Add(echo.HeaderContentType, "application/vnd.oci.image.layer.v1.tar+gzip")
		w.Write([]byte(body)) // nolint: errcheck
	})
	s := httptest.NewServer(mux)

	f := NewClientsFactory()
	cli, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  s.URL,
			TlsVerify: true,
		}})
	assert.NoError(t, err)

	s.Close()

	_, _, err = cli.GetBlob(context.Background(), "library/busybox", dgest)
	assert.Error(t, err)
}

func TestGetBlob4(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	body := "test-busybox"
	hashStr, err := hash.String(body)
	assert.NoError(t, err)
	dgest, err := digest.Parse("sha256:" + hashStr)
	assert.NoError(t, err)
	mux.HandleFunc("/v2/library/busybox/blobs/"+dgest.String(), func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	s := httptest.NewServer(mux)

	f := NewClientsFactory()
	cli, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  s.URL,
			TlsVerify: true,
		}})
	assert.NoError(t, err)

	_, _, err = cli.GetBlob(context.Background(), "library/busybox", dgest)
	assert.Error(t, err)
}

func TestHeadBlob(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	body := "test-busybox"
	hashStr, err := hash.String(body)
	assert.NoError(t, err)
	dgest, err := digest.Parse("sha256:" + hashStr)
	assert.NoError(t, err)
	mux.HandleFunc("/v2/library/busybox/blobs/"+dgest.String(), func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add(consts.ContentDigest, dgest.String())
		w.Header().Add(echo.HeaderContentType, "application/vnd.oci.image.layer.v1.tar+gzip")
		w.Write([]byte(body)) // nolint: errcheck
	})
	s := httptest.NewServer(mux)

	f := NewClientsFactory()
	cli, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  s.URL,
			TlsVerify: true,
		}})
	assert.NoError(t, err)

	descriptor, err := cli.HeadBlob(context.Background(), "library/busybox", dgest)
	assert.NoError(t, err)
	assert.Equal(t, descriptor.Digest.String(), dgest.String())
	assert.Equal(t, descriptor.MediaType, "application/vnd.oci.image.layer.v1.tar+gzip")
	assert.Equal(t, descriptor.Size, int64(len(body)))
}

func TestHeadBlob2(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	body := "test-busybox"
	hashStr, err := hash.String(body)
	assert.NoError(t, err)
	dgest, err := digest.Parse("sha256:" + hashStr)
	assert.NoError(t, err)
	mux.HandleFunc("/v2/library/busybox/blobs/"+dgest.String(), func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add(consts.ContentDigest, "m"+dgest.String())
		w.Header().Add(echo.HeaderContentType, "application/vnd.oci.image.layer.v1.tar+gzip")
		w.Write([]byte(body)) // nolint: errcheck
	})
	s := httptest.NewServer(mux)

	f := NewClientsFactory()
	cli, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  s.URL,
			TlsVerify: true,
		}})
	assert.NoError(t, err)

	_, err = cli.HeadBlob(context.Background(), "library/busybox", dgest)
	assert.Error(t, err)
}

func TestHeadBlob3(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	body := "test-busybox"
	hashStr, err := hash.String(body)
	assert.NoError(t, err)
	dgest, err := digest.Parse("sha256:" + hashStr)
	assert.NoError(t, err)
	mux.HandleFunc("/v2/library/busybox/blobs/"+dgest.String(), func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add(consts.ContentDigest, "m"+dgest.String())
		w.Header().Add(echo.HeaderContentType, "application/vnd.oci.image.layer.v1.tar+gzip")
		w.Write([]byte(body)) // nolint: errcheck
	})
	s := httptest.NewServer(mux)

	f := NewClientsFactory()
	cli, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  s.URL,
			TlsVerify: true,
		}})
	assert.NoError(t, err)

	s.Close()

	_, err = cli.HeadBlob(context.Background(), "library/busybox", dgest)
	assert.Error(t, err)
}

func TestHeadBlob4(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	body := "test-busybox"
	hashStr, err := hash.String(body)
	assert.NoError(t, err)
	dgest, err := digest.Parse("sha256:" + hashStr)
	assert.NoError(t, err)
	mux.HandleFunc("/v2/library/busybox/blobs/"+dgest.String(), func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	s := httptest.NewServer(mux)

	f := NewClientsFactory()
	cli, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  s.URL,
			TlsVerify: true,
		}})
	assert.NoError(t, err)

	_, err = cli.HeadBlob(context.Background(), "library/busybox", dgest)
	assert.Error(t, err)
}
