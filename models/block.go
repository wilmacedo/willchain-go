package models

import (
	"bytes"
	"crypto/sha256"
)

type Block struct {
	Hash         []byte
	Data         []byte
	PreviousHash []byte
	Nonce        int
}

func (block *Block) CalculateHash() {
	data := bytes.Join([][]byte{block.Data, block.PreviousHash}, []byte{})
	hash := sha256.Sum256(data)
	block.Hash = hash[:]
}

func CreateBlock(data string, previousHash []byte) *Block {
	block := &Block{
		Hash:         []byte{},
		Data:         []byte(data),
		PreviousHash: previousHash,
		Nonce:        0,
	}
	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Nonce = nonce
	block.Hash = hash

	return block
}

func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}
