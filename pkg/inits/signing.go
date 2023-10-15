package inits

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/query"
)

func init() {
	inits["signing"] = signing
}

func signing() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Error().Err(err).Msg("Generating RSA private key failed")
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyBytes := pem.EncodeToMemory(privateKeyPEM)

	publicKey := &privateKey.PublicKey

	publicKeyPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(publicKey),
	}
	publicKeyBytes := pem.EncodeToMemory(publicKeyPEM)

	ctx := log.Logger.WithContext(context.Background())

	settingServiceFactory := dao.NewSettingServiceFactory()
	settingService := settingServiceFactory.New()
	_, err = settingService.Get(ctx, consts.SettingSignPrivateKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = query.Q.Transaction(func(tx *query.Query) error {
				settingService := settingServiceFactory.New(tx)
				err = settingService.Save(ctx, consts.SettingSignPrivateKey, privateKeyBytes)
				if err != nil {
					return err
				}
				err = settingService.Save(ctx, consts.SettingSignPublicKey, publicKeyBytes)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				log.Error().Err(err).Msg("Save signing key failed")
				return err
			}
			return nil
		}
		log.Error().Err(err).Msg("Get signing key failed")
		return err
	}
	return nil
}
