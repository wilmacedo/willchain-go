package models

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
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

func (block *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(block)
	Handle(err)

	return result.Bytes()
}

func Deserialize(data []byte) *Block {
	var block *Block
	decoder := gob.NewDecoder(bytes.NewBuffer(data))

	err := decoder.Decode(&block)
	Handle(err)

	return block
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
