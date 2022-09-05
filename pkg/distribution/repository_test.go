package distribution

import (
	"net/http"
	"testing"

	"github.com/distribution/distribution/v3/reference"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	var tran = NewTransport(func(req *http.Request) {
		req.SetBasicAuth("tosone", "8541539655")
	})

	named, err := reference.ParseNamed("library/busybox")
	assert.NoError(t, err)

	repository, err := NewRepository(named, "https://hub.tosone.cn", tran)
	assert.NoError(t, err)

	repository.Manifests()
}
