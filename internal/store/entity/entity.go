package entity

import (
	"database/sql"

	"github.com/uptrace/bun"
)

type shard struct {
	bun.BaseModel `bun:"table:shards"`

	Workchain int32
	Shard     int64
	SeqNo     uint32
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
