package rs

import pb "go.etcd.io/etcd/raft/raftpb"

const (
	DATA_SHARDS     = 3
	PARITY_SHARDS   = 2
	ALL_SHARDS      = DATA_SHARDS + PARITY_SHARDS
	BLOCK_PER_SHARD = 8000
	BLOCK_SIZE      = BLOCK_PER_SHARD * DATA_SHARDS
)

type RsPutEntry struct {
	*encoder
}

type RsGetEntry struct {
	*decoder
}

func EntryEncode(ent pb.Entry) []pb.Entry {
	var rsEnts []pb.Entry
	return rsEnts
}
