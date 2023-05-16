package reader

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimitReader(t *testing.T) {
	originalReader := strings.NewReader("hello world")

	limitReader := LimitReader(originalReader, 5)
	data, err := io.ReadAll(limitReader)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(data))
}
