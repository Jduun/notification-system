package migrations

import (
	"log"
	"notification_system/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
)

func Migrate() {
	m, err := migrate.New(
		"file://migrations",
		config.Cfg.GetDbUrl(),
	)
	if err != nil {
		log.Fatalf("Migration error: %s", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration error: %s", err)
	}
}
