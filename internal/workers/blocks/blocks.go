package blocks

import (
	"context"
	"time"

	zlog "github.com/rs/zerolog/log"

	"tonexplorer/config"
	"tonexplorer/internal/entity"
	"tonexplorer/internal/fetcher"
	"tonexplorer/internal/postgres"
	"tonexplorer/internal/repo/shards"
	"tonexplorer/internal/repo/transactions"
	"tonexplorer/internal/scanner"
	"tonexplorer/pkg/wrapper"
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
	l := zlog.Ctx(ctx).With().
		Str("scope", "blocks worker").
		Logger()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	master, err := worker.startBlock(ctx, worker.cfg)
	if err != nil {
		l.Error().Err(err).Msg("startBlock")
		return wrapper.Wrap(err)
	}

	for {
		select {
		case <-ctx.Done():
			l.Info().Msg("context done")
			return nil
		case <-ticker.C:
			err = worker.db.RunInTx(ctx, nil, func(ctx context.Context, tx postgres.Tx) error {
				l = l.With().Uint32("master", master).Logger()
				l.Info().Msg("start loop")

				shardsRepo := shards.New(tx)
				txsRepo := transactions.New(tx)
				sc := scanner.NewScanner(worker.fetcher, shardsRepo)

				ss, err := sc.Shards(ctx, master)
				if err != nil {
					l.Error().Err(err).Msg("sc.Shards")
					return wrapper.Wrap(err)
				}

				l.Info().Int("shards", len(ss)).Msg("proccesed shard")

				txs, err := sc.Transactions(ctx, ss)
				if err != nil {
					l.Error().Err(err).Msg("sc.Transactions")
					return wrapper.Wrap(err)
				}

				l.Info().Int("txs", len(txs)).Msg("proccesed txs")

				err = shardsRepo.Store(ctx, ss)
				if err != nil {
					l.Error().Err(err).Msg("shardsRepo.Store")
					return wrapper.Wrap(err)
				}

				err = txsRepo.Store(ctx, txs)
				if err != nil {
					l.Error().Err(err).Msg("txsRepo.Store")
					return wrapper.Wrap(err)
				}

				return nil
			})
			if err != nil {
				l.Error().Err(err).Msg("db.RunInTx")
				time.Sleep(5 * time.Second)
				continue
			}

			master++
			l.Info().Msgf("end loop, next master: %d", master)
		}
	}
}

func (worker *W) startBlock(ctx context.Context, cfg *config.C) (uint32, error) {
	l := zlog.Ctx(ctx).With().
		Str("scope", "blocks worker").
		Logger()

	lastVisitedMasterShard, err := worker.shardsRepo.Last(ctx, entity.MasterChain, entity.FirstShard)
	if err != nil {
		l.Error().Err(err).Msg("shardsRepo.Last")
		return 0, wrapper.Wrap(err)
	}

	if lastVisitedMasterShard.SeqNo == 0 {
		return cfg.Genesis, nil
	}

	return lastVisitedMasterShard.SeqNo + 1, nil
}
