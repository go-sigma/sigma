package distribution

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCatalog(t *testing.T) {
	var tran = NewTransport(func(req *http.Request) {
		req.SetBasicAuth("tosone", "8541539655")
	})
	registry, err := NewRegistry("https://hub.tosone.cn", tran)
	assert.NoError(t, err)
	repos, err := registry.Repositories(context.Background())
	assert.NoError(t, err)
	fmt.Println(repos)
}
