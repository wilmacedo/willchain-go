package models

import (
	"bytes"
	"crypto/sha256"
)

type Block struct {
	Hash         []byte
	Data         []byte
	PreviousHash []byte
}

func (block *Block) CalculateHash() {
	info := bytes.Join([][]byte{block.Data, block.PreviousHash}, []byte{})
	hash := sha256.Sum256(info)
	block.Hash = hash[:]
}

func AddBlock(data string, previousHash []byte) *Block {
	block := &Block{
		Hash:         []byte{},
		Data:         []byte(data),
		PreviousHash: previousHash,
	}
	block.CalculateHash()
	return block
}
