package rscode

import (
	"bytes"
	"log"

	"github.com/klauspost/reedsolomon"
	pb "go.etcd.io/etcd/raft/raftpb"
)

const (
	DATA_SHARDS     = 3
	PARITY_SHARDS   = 2
	ALL_SHARDS      = DATA_SHARDS + PARITY_SHARDS
	BLOCK_PER_SHARD = 8000
	BLOCK_SIZE      = BLOCK_PER_SHARD * DATA_SHARDS
)

func EncodeEntry(ent pb.Entry) []pb.Entry {
	ents := []pb.Entry{}
	if ent.Data == nil || len(ent.Data) == 0 || *ent.IndexRS != 0 {
		return ents
	}
	ent_ptr := ent.NextRSEntry
	if ent_ptr != nil {
		for i := 0; i < ALL_SHARDS; i++ {
			ents = append(ents, *ent_ptr)
			ent_ptr = ent_ptr.NextRSEntry
			if ent_ptr == nil {
				return ents
			}
		}
	}
	enc, erre := reedsolomon.New(DATA_SHARDS, PARITY_SHARDS)
	shards, errs := enc.Split(ent.Data)
	if erre == nil && errs == nil && enc.Encode(shards) == nil {
		term := ent.Term
		index := ent.Index
		size := uint32(len(ent.Data))
		ents := make([]pb.Entry, ALL_SHARDS)
		ent_ptr = nil
		for i := ALL_SHARDS - 1; i >= 0; i-- {
			ents[i].Data = shards[i]
			ents[i].Index = index
			ents[i].Term = term
			*ents[i].IndexRS = uint32(i + 1)
			*ents[i].DataSize = size
			ents[i].NextRSEntry = ent_ptr
			ent_ptr = &ents[i]
		}
		log.Printf("Term: %d, Index: %d RS Encoded\n", index, term)
		return ents
	}
	return []pb.Entry{}
}

func DecodeEntries(ents []pb.Entry) pb.Entry {
	if len(ents) < DATA_SHARDS {
		return pb.Entry{}
	} else {
		enc, erre := reedsolomon.New(DATA_SHARDS, PARITY_SHARDS)
		term := ents[0].Term
		index := ents[0].Index
		shards := make([][]byte, ALL_SHARDS)
		for i := range ents {
			if term != ents[i].Term || index != ents[i].Index {
				return pb.Entry{}
			}
			rsIndex := *ents[i].IndexRS
			if rsIndex == 0 {
				return ents[i]
			} else {
				shards[rsIndex-1] = ents[i].Data
			}
		}
		errs := enc.ReconstructData(shards)
		ent := pb.Entry{}
		if erre == nil && errs == nil {
			ent.Term = term
			ent.Index = index
			var buffer bytes.Buffer
			for j := 0; j < DATA_SHARDS; j++ {
				buffer.Write(shards[j])
			}
			ent.Data = buffer.Bytes()[0:*ent.DataSize]
		}
		return ent
	}
}
