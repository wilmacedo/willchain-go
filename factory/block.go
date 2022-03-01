package factory

import (
	"bytes"
	"encoding/gob"

	"github.com/wilmacedo/willchain-go/core"
	"github.com/wilmacedo/willchain-go/factory/merkle"
)

type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PreviousHash []byte
	Nonce        int
}

func (block *Block) HashTransactions() []byte {
	var txHashes [][]byte

	for _, tx := range block.Transactions {
		txHashes = append(txHashes, tx.Serialize())
	}

	tree := merkle.NewMerkleTree(txHashes)

	return tree.RootNode.Data
}

func CreateBlock(txs []*Transaction, previousHash []byte) *Block {
	block := &Block{
		Hash:         []byte{},
		Transactions: txs,
		PreviousHash: previousHash,
		Nonce:        0,
	}
	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Nonce = nonce
	block.Hash = hash

	return block
}

func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

func (block *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(block)
	core.Handle(err)

	return result.Bytes()
}

func Deserialize(data []byte) *Block {
	var block *Block
	decoder := gob.NewDecoder(bytes.NewBuffer(data))

	err := decoder.Decode(&block)
	core.Handle(err)

	return block
}
