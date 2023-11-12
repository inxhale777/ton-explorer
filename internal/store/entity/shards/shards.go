package shards

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"

	"tonexplorer/internal/entity"
	"tonexplorer/internal/postgres"
)

type R struct {
	tx postgres.IDB
}



func New(db postgres.IDB) *R {
	return &R{db}
}

func (r *R) Store(ctx context.Context, ss []entity.Shard) error {
	ssToInsert := make([]shard, 0, len(ss))
	for _, s := range ss {
		ssToInsert = append(ssToInsert, shard{
			Workchain: s.Workchain,
			Shard:     s.Shard,
			SeqNo:     s.SeqNo,
		})
	}

	_, err := r.tx.NewInsert().Model(&ssToInsert).Exec(ctx)
	if err != nil {
		return fmt.Errorf("repo.shards.Store: %w", err)
	}

	return nil
}

func (r *R) Last(ctx context.Context, workchain int32, shardNo int64) (entity.Shard, error) {
	var result shard
	err := r.tx.NewSelect().
		Model(&result).
		Where("workchain = ? and shard = ?", workchain, shardNo).
		Order("seq_no desc").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.Shard{}, nil
		}

		return entity.Shard{}, fmt.Errorf("repo.shards.Last: %w", err)
	}

	return entity.Shard{
		Workchain: result.Workchain,
		Shard:     result.Shard,
		SeqNo:     result.SeqNo,
	}, nil
}
