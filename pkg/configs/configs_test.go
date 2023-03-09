package configs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func errChecker() error {
	return fmt.Errorf("fake error")
}

func noErrChecker() error {
	return nil
}

func TestInitialize(t *testing.T) {
	checkers = make([]checker, 0)
	checkers = append(checkers, noErrChecker, errChecker)
	err := Initialize()
	assert.Error(t, err)
}
