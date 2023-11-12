package entity

const (
	FirstShard  int64 = -9223372036854775808
	MasterChain int32 = -1
)

type Shard struct {
	Workchain int32
	Shard     int64
	SeqNo     uint32
}
