package transactions

import (
	"context"
	"database/sql"
	"fmt"

	"tonexplorer/internal/entity"
	"tonexplorer/internal/postgres"
)

type R struct {
	db postgres.IDB
}

func New(db postgres.IDB) *R {
	return &R{db}
}

func (r *R) Store(ctx context.Context, txs []entity.Transaction) error {
	txsToInsert := make([]entity.transaction, 0, len(txs))
	for _, t := range txs {
		txsToInsert = append(txsToInsert, entity.transaction{
			Hash:        t.Hash,
			Account:     t.Account,
			Success:     t.Success,
			LogicalTime: t.LogicalTime,
			TotalFee:    t.TotalFee,
			Comment: sql.NullString{
				String: t.Comment,
				Valid:  t.Comment != "",
			},
		})
	}

	_, err := r.db.NewInsert().Model(&txsToInsert).Exec(ctx)
	if err != nil {
		return fmt.Errorf("repo.transactions.Store: %w", err)
	}

	return nil
}
