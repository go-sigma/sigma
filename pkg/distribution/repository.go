package distribution

import (
	"net/http"

	"github.com/distribution/distribution/v3/reference"
	rapiv2 "github.com/distribution/distribution/v3/registry/api/v2"
)

// Repository is a named collection of manifests and layers.
type Repository interface {
	// Named returns the name of the repository.
	Named() reference.Named

	// Manifests returns a reference to this repository's manifest service.
	// with the supplied options applied.
	Manifests() (ManifestService, error)

	// Blobs returns a reference to this repository's blob service.
	Blobs() BlobService

	// Tags returns a reference to this repositories tag service
	Tags() TagService
}

type repository struct {
	client *http.Client
	ub     *rapiv2.URLBuilder
	name   reference.Named
}

// NewRepository creates a new Repository for the given name and baseURL.
func NewRepository(name reference.Named, baseURL string, transport http.RoundTripper) (Repository, error) {
	ub, err := rapiv2.NewURLBuilderFromString(baseURL, false)
	if err != nil {
		return nil, err
	}

	return &repository{
		client: &http.Client{
			Transport: transport,
		},
		ub:   ub,
		name: name,
	}, nil
}

// Named returns the name of the repository.
func (r *repository) Named() reference.Named {
	return r.name
}

// Manifests returns a reference to this repository's manifest service.
func (r *repository) Manifests() (ManifestService, error) {
	return &manifest{
		client: r.client,
		ub:     r.ub,
		name:   r.name,
	}, nil
}

// Blobs returns a reference to this repository's blob service.
func (r *repository) Blobs() BlobService {
	return &blob{
		client: r.client,
		ub:     r.ub,
		name:   r.name,
	}
}

// Tags returns a reference to this repositories tag service
func (r *repository) Tags() TagService {
	return &tag{
		client: r.client,
		ub:     r.ub,
		name:   r.name,
	}
}
