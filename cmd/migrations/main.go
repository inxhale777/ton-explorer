package main

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"tonexplorer/config"
)

func main() {
	// only up
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config.New(): %s", err)
	}

	mg, err := migrate.New("file://migrations", cfg.PgDSN)
	if err != nil {
		log.Fatalf("migrate.New(): %s", err)
	}

	err = mg.Up()
	if err != nil {
		log.Fatalf("mg.Up(): %s", err)
	}
}
