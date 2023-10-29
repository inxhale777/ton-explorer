package fetcher

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"log"
	"math/big"
	"tonexplorer/internal/entity"
)

type F struct {
	api            ton.APIClientWrapped
	shardLastSeqno map[string]uint32
}

func New(api ton.APIClientWrapped) *F {
	return &F{api: api, shardLastSeqno: map[string]uint32{}}
}

func (f *F) Run(ctx context.Context) (<-chan entity.Transaction, <-chan error, error) {
	ctx = f.api.Client().StickyContext(ctx)

	master, err := f.api.GetMasterchainInfo(ctx)
	if err != nil {
		log.Fatalf("GetMasterchainInfo() fail: %s", err)
	}

	firstShards, err := f.api.GetBlockShardsInfo(ctx, master)
	if err != nil {
		return nil, nil, fmt.Errorf("GetBlockShardsInfo: %w", err)
	}

	for _, shard := range firstShards {
		f.shardLastSeqno[f.getShardID(shard)] = shard.SeqNo
	}

	stream := make(chan entity.Transaction)
	errCh := make(chan error, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(stream)
				close(errCh)
				return
			default:
				tx, err := f.loop(ctx, master)
				if err != nil {
					errCh <- err
					close(errCh)
					close(stream)
					return
				}

				for _, t := range tx {
					stream <- t
				}

				master, err = f.api.WaitForBlock(master.SeqNo + 1).GetMasterchainInfo(ctx)
				if err != nil {
					errCh <- err
					close(errCh)
					close(stream)
					return
				}
			}
		}
	}()

	return stream, errCh, nil
}

func (f *F) loop(ctx context.Context, current *ton.BlockIDExt) ([]entity.Transaction, error) {
	currentShards, err := f.api.GetBlockShardsInfo(ctx, current)
	if err != nil {
		return nil, err
	}

	// shards in master block may have holes, e.g. shard seqno 2756461, then 2756463, and no 2756462 in master chain
	// thus we need to scan a bit back in case of discovering a hole, till last seen, to fill the misses.
	var newShards []*ton.BlockIDExt
	for _, shard := range currentShards {
		notSeen, err := f.getNotSeenShards(ctx, shard)
		if err != nil {
			return nil, err
		}

		f.shardLastSeqno[f.getShardID(shard)] = shard.SeqNo
		newShards = append(newShards, notSeen...)
	}
	newShards = append(newShards, current)

	var result []entity.Transaction
	// for each shard block getting transactions
	for _, shard := range newShards {
		log.Printf("scanning block %d of shard %x in workchain %d...", shard.SeqNo, uint64(shard.Shard), shard.Workchain)

		var fetchedIDs []ton.TransactionShortInfo
		var after *ton.TransactionID3
		var more = true

		// load all transactions in batches with 100 transactions in each while exists
		for more {
			fetchedIDs, more, err = f.api.WaitForBlock(current.SeqNo).GetBlockTransactionsV2(ctx, shard, 100, after)
			if err != nil {
				return nil, err
			}

			if more {
				// set load offset for next query (pagination)
				after = fetchedIDs[len(fetchedIDs)-1].ID3()
			}

			for _, id := range fetchedIDs {
				// get full transaction by id
				addr := address.NewAddress(0, byte(shard.Workchain), id.Account)
				tx, err := f.api.GetTransaction(ctx, shard, addr, id.LT)
				if err != nil {
					return nil, err
				}

				result = append(result, f.parse(addr, tx))
			}
		}
	}

	return result, nil
}

func (f *F) parse(addr *address.Address, tx *tlb.Transaction) entity.Transaction {
	parsed := entity.Transaction{
		Hash:        hex.EncodeToString(tx.Hash),
		Account:     addr.String(),
		Success:     !bytes.Equal(tx.StateUpdate.OldHash, tx.StateUpdate.NewHash),
		LogicalTime: tx.LT,
		TotalFee:    tx.TotalFees.Coins.String(),
	}

	if tx.IO.In != nil {
		in := new(big.Int)
		if tx.IO.In.MsgType == tlb.MsgTypeInternal {
			in = tx.IO.In.AsInternal().Amount.Nano()
		}

		if in.Cmp(big.NewInt(0)) != 0 {
			parsed.Comment = tx.IO.In.AsInternal().Comment()
		}
	}

	return parsed
}

func (f *F) getNotSeenShards(ctx context.Context, shard *ton.BlockIDExt) ([]*ton.BlockIDExt, error) {
	if no, ok := f.shardLastSeqno[f.getShardID(shard)]; ok && no == shard.SeqNo {
		return nil, nil
	}

	b, err := f.api.GetBlockData(ctx, shard)
	if err != nil {
		return nil, fmt.Errorf("get block data: %w", err)
	}

	parents, err := b.BlockInfo.GetParentBlocks()
	if err != nil {
		return nil, fmt.Errorf("get parent blocks (%d:%x:%d): %w", shard.Workchain, uint64(shard.Shard), shard.Shard, err)
	}

	var result []*ton.BlockIDExt
	for _, parent := range parents {
		ext, err := f.getNotSeenShards(ctx, parent)
		if err != nil {
			return nil, err
		}
		result = append(result, ext...)
	}

	result = append(result, shard)
	return result, nil
}

func (f *F) getShardID(shard *ton.BlockIDExt) string {
	return fmt.Sprintf("%d|%d", shard.Workchain, shard.Shard)
}
