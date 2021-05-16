package rscode

import (
	"bytes"
	"log"

	"github.com/klauspost/reedsolomon"
	serverpb "go.etcd.io/etcd/etcdserver/etcdserverpb"
	pb "go.etcd.io/etcd/raft/raftpb"
)

const (
	DATA_SHARDS     = 3
	PARITY_SHARDS   = 2
	ALL_SHARDS      = DATA_SHARDS + PARITY_SHARDS
	BLOCK_PER_SHARD = 8000
	BLOCK_SIZE      = BLOCK_PER_SHARD * DATA_SHARDS
)

func EncodeEntry(ent pb.Entry) {

	if ent.Data == nil || len(ent.Data) == 0 || ent.NextRSEntry != nil {
		return
	}

	var req, newreq serverpb.InternalRaftRequest
	err := req.Unmarshal(ent.Data)
	if err == nil || req.Put == nil {
		return
	}
	putVal := req.Put.Value

	enc, erre := reedsolomon.New(DATA_SHARDS, PARITY_SHARDS)
	shards, errs := enc.Split(putVal)
	if erre == nil && errs == nil && enc.Encode(shards) == nil {
		ent_ptr := &ent
		term := ent.Term
		index := ent.Index
		var newData []byte
		copy(newData, ent.Data)
		newreq.Unmarshal(newData)
		for _, shard := range shards {
			newreq.Put.Value = shard
			newEntryData, err := newreq.Marshal()
			if err != nil {
				newEntry := &pb.Entry{
					Term:        term,
					Index:       index,
					Data:        newEntryData,
					NextRSEntry: nil,
				}
				ent_ptr.NextRSEntry = newEntry
				ent_ptr = newEntry
			}
		}
		log.Printf("Term: %d, Index: %d RS Encoded\n", index, term)
	}
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
			rsIndex := ents[i].IndexRS
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
			ent.Data = buffer.Bytes()[0:ent.DataSize]
		}
		log.Printf("Term: %d, Index: %d RS Decoded\n", index, term)
		return ent
	}
}
