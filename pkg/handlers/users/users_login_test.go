// Copyright 2023 sigma
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package users

// func TestLogin(t *testing.T) {
// 	logger.SetLevel("debug")
// 	e := echo.New()
// 	validators.Initialize(e)
// 	err := tests.Initialize(t)
// 	assert.NoError(t, err)
// 	err = tests.DB.Init()
// 	assert.NoError(t, err)
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		err = conn.Close()
// 		assert.NoError(t, err)
// 		// err = tests.DB.DeInit()
// 		// assert.NoError(t, err)
// 	}()

// 	viper.Reset()
// 	miniRedis := miniredis.RunT(t)
// 	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

// 	viper.SetDefault("auth.internalUser.password", "internal-sigma")
// 	viper.SetDefault("auth.internalUser.username", "internal-sigma")
// 	viper.SetDefault("auth.admin.password", "Admin@123")
// 	viper.SetDefault("auth.admin.username", "sigma")
// 	err = inits.Initialize()
// 	assert.NoError(t, err)

// 	_, err = handlerNew()
// 	assert.Error(t, err)

// 	viper.SetDefault("auth.jwt.privateKey", privateKeyString)
// 	userHandler, err := handlerNew()
// 	assert.NoError(t, err)

// 	userService := dao.NewUserServiceFactory().New()
// 	userObj, err := userService.GetByUsername(log.Logger.WithContext(context.Background()), "sigma")
// 	assert.NoError(t, err)

// 	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"sigma","password":"Admin@123"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = userHandler.Login(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusOK, c.Response().Status)

// 	// req = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"sigma","password":""}`))
// 	// req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	// rec = httptest.NewRecorder()
// 	// c = e.NewContext(req, rec)
// 	// err = userHandler.Login(c)
// 	// assert.NoError(t, err)
// 	// assert.Equal(t, http.StatusBadRequest, c.Response().Status)
// }

// func TestLoginMockToken(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	logger.SetLevel("debug")
// 	e := echo.New()
// 	validators.Initialize(e)
// 	err := tests.Initialize(t)
// 	assert.NoError(t, err)
// 	err = tests.DB.Init()
// 	assert.NoError(t, err)
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		err = conn.Close()
// 		assert.NoError(t, err)
// 		err = tests.DB.DeInit()
// 		assert.NoError(t, err)
// 	}()

// 	viper.Reset()
// 	miniRedis := miniredis.RunT(t)
// 	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

// 	viper.SetDefault("auth.internalUser.password", "internal-sigma")
// 	viper.SetDefault("auth.internalUser.username", "internal-sigma")
// 	viper.SetDefault("auth.admin.password", "Admin@123")
// 	viper.SetDefault("auth.admin.username", "sigma")
// 	err = inits.Initialize()
// 	assert.NoError(t, err)

// 	var times int
// 	tokenMock := tokenmock.NewMockTokenService(ctrl)
// 	tokenMock.EXPECT().New(gomock.Any(), gomock.Any()).DoAndReturn(func(_ int64, _ time.Duration) (string, error) {
// 		times++
// 		if times == 2 {
// 			return "test", nil
// 		} else {
// 			return "", fmt.Errorf("test")
// 		}
// 	}).Times(3)

// 	viper.SetDefault("auth.jwt.privateKey", privateKeyString)
// 	userHandler, err := handlerNew(inject{tokenService: tokenMock})
// 	assert.NoError(t, err)

// 	userService := dao.NewUserServiceFactory().New()
// 	userObj, err := userService.GetByUsername(log.Logger.WithContext(context.Background()), "sigma")
// 	assert.NoError(t, err)

// 	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"sigma","password":"Admin@123"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = userHandler.Login(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)

// 	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"sigma","password":"Admin@123"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = userHandler.Login(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
// }

// func TestLoginMockPassword(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	// passwordMock := passwordmock.NewMockPassword(ctrl)
// 	// passwordMock.EXPECT().Verify(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ string) bool {
// 	// 	return false
// 	// }).Times(1)

// 	logger.SetLevel("debug")
// 	e := echo.New()
// 	validators.Initialize(e)
// 	err := tests.Initialize(t)
// 	assert.NoError(t, err)
// 	err = tests.DB.Init()
// 	assert.NoError(t, err)
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		err = conn.Close()
// 		assert.NoError(t, err)
// 		err = tests.DB.DeInit()
// 		assert.NoError(t, err)
// 	}()

// 	viper.Reset()
// 	miniRedis := miniredis.RunT(t)
// 	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

// 	viper.SetDefault("auth.internalUser.password", "internal-sigma")
// 	viper.SetDefault("auth.internalUser.username", "internal-sigma")
// 	viper.SetDefault("auth.admin.password", "Admin@123")
// 	viper.SetDefault("auth.admin.username", "sigma")
// 	err = inits.Initialize()
// 	assert.NoError(t, err)

// 	viper.SetDefault("auth.jwt.privateKey", privateKeyString)
// 	userHandler, err := handlerNew(inject{})
// 	assert.NoError(t, err)

// 	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"sigma","password":"Admin@123"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	err = userHandler.Login(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusUnauthorized, c.Response().Status)
// }

// func TestLoginMockDAO(t *testing.T) {
// 	viper.Reset()
// 	logger.SetLevel("debug")
// 	e := echo.New()
// 	e.HideBanner = true
// 	e.HidePort = true
// 	validators.Initialize(e)
// 	err := tests.Initialize(t)
// 	assert.NoError(t, err)
// 	err = tests.DB.Init()
// 	assert.NoError(t, err)
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		err = conn.Close()
// 		assert.NoError(t, err)
// 		err = tests.DB.DeInit()
// 		assert.NoError(t, err)
// 	}()

// 	viper.SetDefault("auth.internalUser.password", "internal-sigma")
// 	viper.SetDefault("auth.internalUser.username", "internal-sigma")
// 	viper.SetDefault("auth.admin.password", "Admin@123")
// 	viper.SetDefault("auth.admin.username", "sigma")
// 	viper.SetDefault("auth.jwt.privateKey", privateKeyString)

// 	miniRedis := miniredis.RunT(t)
// 	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

// 	err = inits.Initialize()
// 	assert.NoError(t, err)

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	// daoMockUserService := daomock.NewMockUserService(ctrl)
// 	// daoMockUserService.EXPECT().GetByUsername(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ string) (*models.User, error) {
// 	// 	return nil, fmt.Errorf("test")
// 	// }).Times(1)

// 	// daoMockUserServiceFactory := daomock.NewMockUserServiceFactory(ctrl)
// 	// daoMockUserServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.UserService {
// 	// 	return daoMockUserService
// 	// }).Times(1)

// 	viper.SetDefault("auth.jwt.privateKey", privateKeyString)
// 	userHandler, err := handlerNew(inject{})
// 	assert.NoError(t, err)

// 	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"test","password":"123498712311Aa!"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	err = userHandler.Login(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusUnauthorized, c.Response().Status)
// }
