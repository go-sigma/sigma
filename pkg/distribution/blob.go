package distribution

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	rapiv2 "github.com/distribution/distribution/v3/registry/api/v2"
	"github.com/opencontainers/go-digest"
)

type Mount struct {
	From   reference.Canonical
	Digest digest.Digest
}

// BlobService provides access to blobs by digest that are part of repositories
type BlobService interface {
	Stat(ctx context.Context, dgst digest.Digest) (*distribution.Descriptor, error)
	Get(ctx context.Context, dgst digest.Digest) ([]byte, error)
	Open(ctx context.Context, dgst digest.Digest) (io.ReadCloser, error)
	Put(ctx context.Context, mediaType string, p []byte) (distribution.Descriptor, error)
	Create(ctx context.Context) (BlobWriter, error)
	Resume(ctx context.Context, id string) (BlobWriter, error)
}

type blob struct {
	client *http.Client
	ub     *rapiv2.URLBuilder
	name   reference.Named
}

// Stat returns the descriptor for the blob identified by the digest.
func (b *blob) Stat(ctx context.Context, dgst digest.Digest) (*distribution.Descriptor, error) {
	ref, err := reference.WithDigest(b.name, dgst)
	if err != nil {
		return nil, err
	}
	u, err := b.ub.BuildBlobURL(ref)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // nolint: errcheck
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	dgstHeader := resp.Header.Get("Docker-Content-Digest")
	if dgstHeader == "" {
		return nil, fmt.Errorf("missing digest header")
	}
	lengthHeader := resp.Header.Get("Content-Length")
	if lengthHeader == "" {
		return nil, fmt.Errorf("missing content-length header for request: %s", u)
	}
	length, err := strconv.ParseInt(lengthHeader, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing content-length: %v", err)
	}
	return &distribution.Descriptor{
		MediaType: resp.Header.Get("Content-Type"),
		Size:      length,
		Digest:    dgst,
	}, nil
}

// Get returns the entire blob identified by the digest.
func (b *blob) Get(ctx context.Context, dgst digest.Digest) ([]byte, error) {
	reader, err := b.Open(ctx, dgst)
	if err != nil {
		return nil, err
	}
	defer reader.Close() // nolint: errcheck
	return io.ReadAll(reader)
}

// Open returns a ReadCloser for the blob identified by the digest.
func (b *blob) Open(ctx context.Context, dgst digest.Digest) (io.ReadCloser, error) {
	ref, err := reference.WithDigest(b.name, dgst)
	if err != nil {
		return nil, err
	}
	u, err := b.ub.BuildBlobURL(ref)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // nolint: errcheck
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	return resp.Body, nil
}

// Put inserts the contents of p into the blob store, returning the descriptor
// of the content. The descriptor's digest may be different than the digest of
// p if the blob store performs data validation on write.
func (b *blob) Put(ctx context.Context, mediaType string, p []byte) (distribution.Descriptor, error) {
	return distribution.Descriptor{}, fmt.Errorf("not implemented")
}

// Create creates a new blob writer with a randomly generated ID.
func (b *blob) Create(ctx context.Context) (BlobWriter, error) {
	panic("not implemented")
}

// Resume creates a new blob writer with the given ID. If a writer with the
// given ID does not exist, an error will be returned.
func (b *blob) Resume(ctx context.Context, id string) (BlobWriter, error) {
	panic("not implemented")
}
