package distribution

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/distribution/distribution/v3/reference"
	"github.com/stretchr/testify/assert"
)

func TestAllTags(t *testing.T) {
	var tran = NewTransport(func(req *http.Request) {
		req.SetBasicAuth("tosone", "8541539655")
	})

	named, err := reference.WithName("library/busybox")
	assert.NoError(t, err)

	repository, err := NewRepository(named, "https://hub.tosone.cn", tran)
	assert.NoError(t, err)
	tags, err := repository.Tags().All(context.Background())
	assert.NoError(t, err)
	fmt.Println(tags)
}
