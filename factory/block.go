package factory

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"

	"github.com/wilmacedo/willchain-go/core"
	"github.com/wilmacedo/willchain-go/models"
)

type Block struct {
	Hash         []byte
	Transactions []*models.Transaction
	PreviousHash []byte
	Nonce        int
}

func (block *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range block.Transactions {
		txHashes = append(txHashes, tx.Hash)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

func CreateBlock(txs []*models.Transaction, previousHash []byte) *Block {
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

func Genesis(coinbase *models.Transaction) *Block {
	return CreateBlock([]*models.Transaction{coinbase}, []byte{})
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
