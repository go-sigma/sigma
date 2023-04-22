package inits

import (
	"context"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/logger"
	"github.com/ximager/ximager/pkg/tests"
	"github.com/ximager/ximager/pkg/utils/password"
)

func TestInitInternalUser(t *testing.T) {
	logger.SetLevel("debug")
	err := tests.Initialize()
	assert.NoError(t, err)
	err = tests.DB.Init()
	assert.NoError(t, err)
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		err = conn.Close()
		assert.NoError(t, err)
		err = tests.DB.DeInit()
		assert.NoError(t, err)
	}()

	err = initUser()
	assert.Error(t, err)

	viper.SetDefault("auth.internalUser.password", "internal-ximager")
	err = initUser()
	assert.Error(t, err)

	viper.SetDefault("auth.internalUser.username", "internal-ximager")
	err = initUser()
	assert.Error(t, err)
}

func TestInitAdminUser1(t *testing.T) {
	logger.SetLevel("debug")
	err := tests.Initialize()
	assert.NoError(t, err)
	err = tests.DB.Init()
	assert.NoError(t, err)
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		err = conn.Close()
		assert.NoError(t, err)
		err = tests.DB.DeInit()
		assert.NoError(t, err)
	}()

	viper.SetDefault("auth.internalUser.password", "internal-ximager")
	viper.SetDefault("auth.internalUser.username", "internal-ximager")

	viper.SetDefault("auth.admin.password", "ximager")
	err = initUser()
	assert.Error(t, err)
}

func TestInitAdminUser2(t *testing.T) {
	logger.SetLevel("debug")
	err := tests.Initialize()
	assert.NoError(t, err)
	err = tests.DB.Init()
	assert.NoError(t, err)
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		err = conn.Close()
		assert.NoError(t, err)
		err = tests.DB.DeInit()
		assert.NoError(t, err)
	}()

	viper.SetDefault("auth.internalUser.password", "internal-ximager")
	viper.SetDefault("auth.internalUser.username", "internal-ximager")
	viper.SetDefault("auth.admin.password", "ximager")
	viper.SetDefault("auth.admin.username", "ximager")
	viper.SetDefault("auth.admin.email", "ximager@gmail.com")
	err = initUser()
	assert.NoError(t, err)

	userService := dao.NewUserService()
	passwordService := password.New()

	count, err := userService.Count(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, count, int64(2))

	ctx := context.Background()
	user, err := userService.GetByUsername(ctx, "ximager")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.True(t, passwordService.Verify("ximager", user.Password))
	assert.Equal(t, user.Role, "root")

	user, err = userService.GetByUsername(ctx, "internal-ximager")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.True(t, passwordService.Verify("internal-ximager", user.Password))
	assert.Equal(t, user.Role, "root")
}
