package postgres

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/schema"
)

type P struct {
	*bun.DB
}

type IDB interface {
	Dialect() schema.Dialect
	NewSelect() *bun.SelectQuery
	NewInsert() *bun.InsertQuery
	NewUpdate() *bun.UpdateQuery
}

type DB interface {
	bun.IConn
	IDB
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	RunInTx(ctx context.Context, opts *sql.TxOptions, f func(context.Context, Tx) error) error
}

type Tx interface {
	bun.IConn
	Dialect() schema.Dialect
	NewSelect() *bun.SelectQuery
	NewInsert() *bun.InsertQuery
	NewUpdate() *bun.UpdateQuery
	Commit() error
	Rollback() error
}

func New(dsn string) (*P, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	return &P{db}, nil
}

func (p *P) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	return p.DB.BeginTx(ctx, opts)
}

func (p *P) RunInTx(ctx context.Context, opts *sql.TxOptions, fn func(context.Context, Tx) error) error {
	tx, err := p.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	var done bool
	defer func() {
		if !done {
			_ = tx.Rollback()
		}
	}()

	if err := fn(ctx, tx); err != nil {
		return err
	}

	done = true
	return tx.Commit()
}

func (p *P) InsertTxsAndShard(txs, shard []interface{}) error {
	// begin tx....

	// insert

	// commit
}