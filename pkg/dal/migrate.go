package dal

import (
	"embed"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var fs embed.FS

func migrateMysql(dsn string) error {
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, fmt.Sprintf("mysql://%s", dsn))
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	version, dirty, err := m.Version()
	if err != nil {
		return err
	}
	log.Info().Uint("version", version).Bool("dirty", dirty).Msg("Migrate database")
	return nil
}

func migratePostgres(dsn string) error {
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, fmt.Sprintf("postgres://%s", dsn))
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	version, dirty, err := m.Version()
	if err != nil {
		return err
	}
	log.Info().Uint("version", version).Bool("dirty", dirty).Msg("Migrate database")
	return nil
}
