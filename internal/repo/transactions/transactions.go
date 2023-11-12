package transactions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"

	"tonexplorer/internal/entity"
	"tonexplorer/internal/postgres"
)

type R struct {
	db postgres.IDB
}

type transaction struct {
	bun.BaseModel `bun:"table:transactions"`

	Hash        string
	Account     string
	Success     bool
	LogicalTime uint64
	TotalFee    string
	Comment     sql.NullString
}

func New(db postgres.IDB) *R {
	return &R{db}
}

func (r *R) Store(ctx context.Context, txs []entity.Transaction) error {
	txsToInsert := make([]transaction, 0, len(txs))
	for _, t := range txs {
		txsToInsert = append(txsToInsert, transaction{
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
