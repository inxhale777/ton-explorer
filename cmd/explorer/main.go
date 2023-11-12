package main

import (
	"context"
	"log"
	"tonexplorer/config"
	"tonexplorer/internal/fetcher"
	"tonexplorer/internal/postgres"
	"tonexplorer/internal/repo/shards"
	"tonexplorer/internal/repo/transactions"
	"tonexplorer/internal/tclient"
	"tonexplorer/internal/workers/blocks"
	server2 "tonexplorer/internal/workers/server"
)

func main() {
	ctx := context.Background()

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config.New() fail: %s", err)
	}

	api, err := tclient.New()
	if err != nil {
		log.Fatalf("API() fail: %s", err)
	}

	db, err := postgres.New(cfg.PgDSN)
	if err != nil {
		log.Fatal(err)
		return
	}

	f := fetcher.New(api)
	shardsRepo := shards.New(db)
	transactionsRepo := transactions.New(db)

	go func() {
		w := blocks.New(cfg, f, shardsRepo, db)
		err := w.Run(ctx)
		if err != nil {
			log.Fatalf("blocks worker exit: %s", err)
		}
	}()

	go func() {
		s := server2.New()
		err := s.Run(ctx, shardsRepo, transactionsRepo)
		if err != nil {
			log.Fatalf("server worker exit: %s", err)
		}
	}()

	select {}
}
