package server

import (
	"context"

	"tonexplorer/internal/api"
	"tonexplorer/internal/repo/shards"
	"tonexplorer/internal/repo/transactions"
)

type W struct{}

func New() *W {
	return &W{}
}

func (worker *W) Run(ctx context.Context, shardsRepo *shards.R, transactionsRepo *transactions.R, endpoint string) error {
	a := api.New(shardsRepo, transactionsRepo)
	router := a.Init()

	return router.Start(endpoint)
}
