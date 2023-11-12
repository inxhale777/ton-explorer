CREATE TABLE shards
(
    workchain integer not null,
    shard     bigint  not null,
    seq_no    integer not null,
    unique (workchain, shard, seq_no)
);
