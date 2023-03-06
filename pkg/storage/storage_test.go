package storage

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

type dummyFactory struct{}

func (dummyFactory) New() (StorageDriver, error) {
	return nil, nil
}

type dummyFactoryError struct{}

func (dummyFactoryError) New() (StorageDriver, error) {
	return nil, fmt.Errorf("dummy error")
}

func TestRegisterDriverFactory(t *testing.T) {
	driverFactories = make(map[string]Factory)

	err := RegisterDriverFactory("dummy", &dummyFactory{})
	assert.NoError(t, err)

	err = RegisterDriverFactory("dummy", &dummyFactory{})
	assert.Error(t, err)
}

func TestInitialize(t *testing.T) {
	driverFactories = make(map[string]Factory)

	err := RegisterDriverFactory("dummy", &dummyFactory{})
	assert.NoError(t, err)

	viper.SetDefault("storage.type", "dummy")
	err = Initialize()
	assert.NoError(t, err)

	viper.SetDefault("storage.type", "fake")
	err = Initialize()
	assert.Error(t, err)

	err = RegisterDriverFactory("dummy-error", &dummyFactoryError{})
	assert.NoError(t, err)

	viper.SetDefault("storage.type", "dummy-error")
	err = Initialize()
	assert.Error(t, err)
}
