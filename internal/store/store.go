package store


type TonSharer interface {
	// Last(ctx context.Context, entity.MasterChain, entity.FirstShard)
	InsertTxsAndShard(txs, shard []interface{}) error
}
