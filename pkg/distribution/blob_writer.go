package distribution

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/distribution/distribution/v3"
	"github.com/rs/zerolog/log"
)

// BlobWriter provides a handle for inserting data into a blob store.
// Instances should be obtained from BlobWriteService.Writer and
// BlobWriteService.Resume. If supported by the store, a writer can be
// recovered with the id.
type BlobWriter interface {
	io.WriteCloser
	io.ReaderFrom

	// Size returns the number of bytes written to this blob.
	Size() int64

	// ID returns the identifier for this writer. The ID can be used with the
	// Blob service to later resume the write.
	ID() string

	// StartedAt returns the time this blob write was started.
	StartedAt() time.Time

	// Commit completes the blob writer process. The content is verified
	// against the provided provisional descriptor, which may result in an
	// error. Depending on the implementation, written data may be validated
	// against the provisional descriptor fields. If MediaType is not present,
	// the implementation may reject the commit or assign "application/octet-
	// stream" to the blob. The returned descriptor may have a different
	// digest depending on the blob store, referred to as the canonical
	// descriptor.
	Commit(ctx context.Context, provisional distribution.Descriptor) (distribution.Descriptor, error)

	// Cancel ends the blob write without storing any data and frees any
	// associated resources. Any data written thus far will be lost. Cancel
	// implementations should allow multiple calls even after a commit that
	// result in a no-op. This allows use of Cancel in a defer statement,
	// increasing the assurance that it is correctly called.
	Cancel(ctx context.Context) error
}

type blobWriter struct {
	ctx context.Context

	blobService BlobService
	client      *http.Client

	uuid      string
	startedAt time.Time

	location string // always the last value of the location header.
	offset   int64
	closed   bool
}

func (hbu *blobWriter) ReadFrom(r io.Reader) (n int64, err error) {
	req, err := http.NewRequestWithContext(hbu.ctx, http.MethodPatch, hbu.location, io.NopCloser(r))
	if err != nil {
		return 0, err
	}
	defer req.Body.Close()

	resp, err := hbu.client.Do(req)
	if err != nil {
		return 0, err
	}

	if !SuccessStatus(resp.StatusCode) {
		log.Error().Int("code", resp.StatusCode).Msg("unexpected status code")
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	hbu.uuid = resp.Header.Get("Docker-Upload-UUID")
	hbu.location, err = sanitizeLocation(resp.Header.Get("Location"), hbu.location)
	if err != nil {
		return 0, err
	}
	rng := resp.Header.Get("Range")
	var start, end int64
	if n, err := fmt.Sscanf(rng, "%d-%d", &start, &end); err != nil {
		return 0, err
	} else if n != 2 || end < start {
		return 0, fmt.Errorf("bad range format: %s", rng)
	}

	hbu.offset += end - start + 1
	return (end - start + 1), nil
}

func (hbu *blobWriter) Write(p []byte) (n int, err error) {
	req, err := http.NewRequestWithContext(hbu.ctx, http.MethodPatch, hbu.location, bytes.NewReader(p))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Range", fmt.Sprintf("%d-%d", hbu.offset, hbu.offset+int64(len(p)-1)))
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(p)))
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := hbu.client.Do(req)
	if err != nil {
		return 0, err
	}

	if !SuccessStatus(resp.StatusCode) {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	hbu.uuid = resp.Header.Get("Docker-Upload-UUID")
	hbu.location, err = sanitizeLocation(resp.Header.Get("Location"), hbu.location)
	if err != nil {
		return 0, err
	}
	rng := resp.Header.Get("Range")
	var start, end int
	if n, err := fmt.Sscanf(rng, "%d-%d", &start, &end); err != nil {
		return 0, err
	} else if n != 2 || end < start {
		return 0, fmt.Errorf("bad range format: %s", rng)
	}

	hbu.offset += int64(end - start + 1)
	return (end - start + 1), nil
}

func (hbu *blobWriter) Size() int64 {
	return hbu.offset
}

func (hbu *blobWriter) ID() string {
	return hbu.uuid
}

func (hbu *blobWriter) StartedAt() time.Time {
	return hbu.startedAt
}

func (hbu *blobWriter) Commit(ctx context.Context, desc distribution.Descriptor) (*distribution.Descriptor, error) {
	req, err := http.NewRequestWithContext(hbu.ctx, http.MethodPut, hbu.location, nil)
	if err != nil {
		return nil, err
	}

	values := req.URL.Query()
	values.Set("digest", desc.Digest.String())
	req.URL.RawQuery = values.Encode()

	resp, err := hbu.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !SuccessStatus(resp.StatusCode) {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	return hbu.blobService.Stat(ctx, desc.Digest)
}

func (hbu *blobWriter) Cancel(ctx context.Context) error {
	req, err := http.NewRequestWithContext(hbu.ctx, http.MethodDelete, hbu.location, nil)
	if err != nil {
		return err
	}
	resp, err := hbu.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || SuccessStatus(resp.StatusCode) {
		return nil
	}
	return fmt.Errorf("unexpected status code %d", resp.StatusCode)
}

func (hbu *blobWriter) Close() error {
	hbu.closed = true
	return nil
}

func sanitizeLocation(location, base string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	locationURL, err := url.Parse(location)
	if err != nil {
		return "", err
	}

	return baseURL.ResolveReference(locationURL).String(), nil
}

// SuccessStatus returns true if the argument is a successful HTTP response
// code (in the range 200 - 399 inclusive).
func SuccessStatus(status int) bool {
	return status >= 200 && status <= 399
}
