package blocks

import (
	"context"
	"log"
	"time"

	"tonexplorer/config"
	"tonexplorer/internal/entity"
	"tonexplorer/internal/fetcher"
	"tonexplorer/internal/postgres"
	"tonexplorer/internal/repo/shards"
	"tonexplorer/internal/repo/transactions"
	"tonexplorer/internal/scanner"
)

type W struct {
	cfg        *config.C
	fetcher    *fetcher.F
	shardsRepo *shards.R
	db         postgres.DB
}

func New(cfg *config.C, fetcher *fetcher.F, shardsRepo *shards.R, db postgres.DB) *W {
	return &W{cfg, fetcher, shardsRepo, db}
}

func (worker *W) Run(ctx context.Context) error {
	const trace = "blocks worker"

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	master, err := worker.startBlock(ctx, worker.cfg)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err = worker.db.RunInTx(ctx, nil, func(ctx context.Context, tx postgres.Tx) error {
				log.Printf("%s: is about to scan %d master block", trace, master)

				shardsRepo := shards.New(tx)
				txsRepo := transactions.New(tx)
				sc := scanner.New(worker.fetcher, shardsRepo)

				ss, err := sc.Shards(ctx, master)
				if err != nil {
					return err
				}

				log.Printf("%s: proccesed %d shard from master block %d", trace, len(ss), master)

				txs, err := sc.Transactions(ctx, ss)
				if err != nil {
					return err
				}

				log.Printf("%s: proccesed %d txs from master block %d", trace, len(txs), master)

				err = shardsRepo.Store(ctx, ss)
				if err != nil {
					return err
				}

				err = txsRepo.Store(ctx, txs)
				if err != nil {
					return err
				}

				return nil
			})
			if err != nil {
				log.Printf("%s: loop failed: %s", trace, err)
				time.Sleep(5 * time.Second)
				continue
			}

			master++
		}
	}
}

func (worker *W) startBlock(ctx context.Context, cfg *config.C) (uint32, error) {
	lastVisitedMasterShard, err := worker.shardsRepo.Last(ctx, entity.MasterChain, entity.FirstShard)
	if err != nil {
		return 0, err
	}

	if lastVisitedMasterShard.SeqNo == 0 {
		return cfg.Genesis, nil
	}

	return lastVisitedMasterShard.SeqNo + 1, nil
}
