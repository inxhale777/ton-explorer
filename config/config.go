package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type C struct {
	Endpoint string `yaml:"endpoint" env:"ENDPOINT"`
	PgDSN    string `yaml:"pg_dsn" env:"PG_DSN"`
	Genesis  uint32 `yaml:"genesis" env:"GENESIS"`
}

func New() (*C, error) {
	cfg := &C{}
	config := os.Getenv("CONFIG_PATH")
	if config == "" {
		config = "./config/config.yml"
	}

	err := cleanenv.ReadConfig(config, cfg)
	if err != nil {
		return nil, fmt.Errorf("config.New: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
