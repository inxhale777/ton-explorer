package fetcher

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton"

	"tonexplorer/internal/entity"
)

type F struct {
	api ton.APIClientWrapped
}

func New(api ton.APIClientWrapped) *F {
	return &F{api}
}

func (f *F) Shards(ctx context.Context, masterSeqNo uint32) ([]entity.Shard, error) {
	const trace = "fetcher.Shards"

	blk, err := f.api.LookupBlock(ctx, entity.MasterChain, entity.FirstShard, masterSeqNo)
	if err != nil {
		return nil, fmt.Errorf("%s: LookupBlock: %w", trace, err)
	}

	shards, err := f.api.GetBlockShardsInfo(ctx, blk)
	if err != nil {
		return nil, fmt.Errorf("%s: GetBlockShardsInfo: %w", trace, err)
	}

	result := make([]entity.Shard, 0, len(shards)+1)
	result = append(result, entity.Shard{
		Workchain: blk.Workchain,
		Shard:     blk.Shard,
		SeqNo:     blk.SeqNo,
	})

	for _, s := range shards {
		result = append(result, entity.Shard{
			Workchain: s.Workchain,
			Shard:     s.Shard,
			SeqNo:     s.SeqNo,
		})
	}

	return result, nil
}

func (f *F) Transactions(ctx context.Context, shard entity.Shard) ([]entity.Transaction, error) {
	const trace = "fetcher.Transactions"

	blk, err := f.api.LookupBlock(ctx, shard.Workchain, shard.Shard, shard.SeqNo)
	if err != nil {
		return nil, fmt.Errorf("%s: LookupBlock: %w", trace, err)
	}

	var result []entity.Transaction
	var after *ton.TransactionID3
	more := true
	for more {
		var ids []ton.TransactionShortInfo
		ids, more, err = f.api.GetBlockTransactionsV2(ctx, blk, 100, after)
		if err != nil {
			return nil, fmt.Errorf("%s: GetBlockTransactionsV2: %w", trace, err)
		}

		if len(ids) > 0 {
			after = ids[len(ids)-1].ID3()
		}

		for _, i := range ids {
			addr := address.NewAddress(0, byte(blk.Workchain), i.Account)
			tx, err := f.api.GetTransaction(ctx, blk, addr, i.LT)
			if err != nil {
				return nil, fmt.Errorf("%s: GetTransaction: %w", trace, err)
			}

			result = append(result, entity.Transaction{
				Hash:        hex.EncodeToString(tx.Hash),
				Account:     addr.String(),
				Success:     !bytes.Equal(tx.StateUpdate.OldHash, tx.StateUpdate.NewHash),
				LogicalTime: tx.LT,
				TotalFee:    tx.TotalFees.Coins.String(),
			})
		}
	}

	return result, nil
}
