package scanner

import (
	"context"
	"fmt"
	"tonexplorer/internal/entity"
	"tonexplorer/internal/fetcher"
	"tonexplorer/internal/repo/shards"
	"tonexplorer/pkg/wrapper"

	zlog "github.com/rs/zerolog/log"
)

type Scanner struct {
	fetcher *fetcher.F
	shards  *shards.R
}

func NewScanner(fetcher *fetcher.F, shards *shards.R) *S {
	return &Scanner{fetcher, shards}
}

// Shards find and return all unvisited shards from master block
func (s *Scanner) Shards(ctx context.Context, masterSeqNo uint32) ([]entity.Shard, error) {
	l := zlog.Ctx(ctx).With().
		Str("scope", "scanner.Shards").
		Uint32("master_seq_no", masterSeqNo).Logger()

	allShardFromMasterBlk, err := s.fetcher.Shards(ctx, masterSeqNo)
	if err != nil {
		l.Error().Err(err).Msg("fetcher.Shards")
		return nil, wrapper.Wrap(err)
	}

	var result []entity.Shard
	for _, shard := range allShardFromMasterBlk {
		// find number of last visited shard on that chain
		lastProcessedShard, err := s.shards.Last(ctx, shard.Workchain, shard.Shard)
		if err != nil {
			l.Error().Err(err).Msg("shards.Last")
			return nil, wrapper.Wrap(err)
		}

		var nextShard uint32
		if lastProcessedShard.SeqNo == shard.SeqNo {
			// have been here already
			continue
		} else if lastProcessedShard.SeqNo == 0 {
			// haven't processed any shard from that chain
			nextShard = shard.SeqNo
		} else {
			// lets process next shard right after the last stored one
			nextShard = lastProcessedShard.SeqNo + 1
		}

		// iterate from last visited shard to current shard in specified master block
		for shard.SeqNo >= nextShard {
			result = append(result, entity.Shard{
				Workchain: shard.Workchain,
				Shard:     shard.Shard,
				SeqNo:     nextShard,
			})

			nextShard++
		}
	}
	l.Info().Int("shards_count", len(result)).Msg("found shards")

	return result, err
}

// Transactions fetch all transactions from all provided shards
func (scanner *S) Transactions(ctx context.Context, shards []entity.Shard) ([]entity.Transaction, error) {
	const trace = "scanner.Transactions"

	var result []entity.Transaction
	for _, s := range shards {
		txs, err := scanner.fetcher.Transactions(ctx, s)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", trace, err)
		}

		result = append(result, txs...)
	}

	return result, nil
}
