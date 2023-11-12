package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"tonexplorer/config"
	"tonexplorer/internal/fetcher"
	"tonexplorer/internal/postgres"
	"tonexplorer/internal/repo/shards"
	"tonexplorer/internal/repo/transactions"
	"tonexplorer/internal/tclient"
	"tonexplorer/internal/workers/blocks"
	server2 "tonexplorer/internal/workers/server"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

var (
	shaCommit string = "dev"
)

func main() {
	ctx := context.Background()

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config.New() fail: %s", err)
	}

	// zerolog

	debug := os.Getenv("DEBUG")
	if debug != "" {
		var lvl zerolog.Level
		switch debug {
		case "1":
			lvl = zerolog.InfoLevel
		case "2":
			lvl = zerolog.DebugLevel
		case "3":
			lvl = zerolog.TraceLevel
		}

		fmt.Printf("Zerolog level: %s\n", lvl.String())

		writer := zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.TimeOnly,
		}
		zlog.Logger = zerolog.New(writer).
			Level(lvl).
			With().
			Timestamp().
			Caller().
			Logger()

	} else {
		// json logger for production
		zlog.Logger = zerolog.New(os.Stderr).
			Level(zerolog.InfoLevel).
			With().
			Timestamp().
			Caller().
			Logger()

		zlog.Logger.Info().Msgf("Starting mill with commit %s", shaCommit)
	}
	l := zlog.Logger
	ctx = zlog.Logger.WithContext(ctx)

	api, err := tclient.NewClient()
	if err != nil {
		l.Fatal().Err(err).Msg("tclient.NewClient")
	}

	db, err := postgres.New(cfg.PgDSN)
	if err != nil {
		l.Fatal().Err(err).Msg("postgres.New")
	}

	f := fetcher.NewFetcher(api)
	shardsRepo := shards.New(db)
	transactionsRepo := transactions.New(db)

	// Start worker
	go func() {
		w := blocks.New(cfg, f, shardsRepo, db)
		err := w.Run(ctx)
		if err != nil {
			l.Fatal().Err(err).Msg("blocks worker exit")
		}
	}()

	// start server
	go func() {
		s := server2.New()
		err := s.Run(ctx, shardsRepo, transactionsRepo, cfg.Endpoint)
		if err != nil {
			l.Fatal().Err(err).Msg("server worker exit")
		}
	}()

	select {}
}
