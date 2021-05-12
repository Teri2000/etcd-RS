package rs

import (
	"io"

	"github.com/klauspost/reedsolomon"
)

type encoder struct {
	writer []io.Writer
	enc    reedsolomon.Encoder
	cache  []byte
}

func newEncoder(writers []io.Writer) *encoder {
	enc, _ := reedsolomon.New(DATA_SHARDS, PARITY_SHARDS)
	return &encoder{writers, enc, nil}
}
