package models

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	dbPath = "./tmp/blocks"
)

type Blockchain struct {
	LastHash []byte
	Database *leveldb.DB
}

type Iterator struct {
	CurrentHash []byte
	Database    *leveldb.DB
}

func InitBlockchain() *Blockchain {
	var lastHash []byte

	db, err := leveldb.OpenFile(dbPath, nil)
	Handle(err)

	if _, err := db.Get([]byte("lh"), nil); err == leveldb.ErrNotFound {
		fmt.Println("No existing blockchain found")
		genesis := Genesis()
		fmt.Println("Genesis proved")

		err := db.Put(genesis.Hash, genesis.Serialize(), nil)
		Handle(err)

		err = db.Put([]byte("lh"), genesis.Hash, nil)
		Handle(err)

		lastHash = genesis.Hash
	} else {
		lastHash, err = db.Get([]byte("lh"), nil)
		Handle(err)
	}

	chain := &Blockchain{
		LastHash: lastHash,
		Database: db,
	}

	return chain
}

func (chain *Blockchain) AddBlock(data string) {
	lastHash, err := chain.Database.Get([]byte("lh"), nil)
	Handle(err)

	newBlock := CreateBlock(data, lastHash)

	err = chain.Database.Put(newBlock.Hash, newBlock.Serialize(), nil)
	Handle(err)

	err = chain.Database.Put([]byte("lh"), newBlock.Hash, nil)
	Handle(err)

	chain.LastHash = newBlock.Hash
}

func (chain *Blockchain) Iterator() *Iterator {
	iter := &Iterator{
		CurrentHash: chain.LastHash,
		Database:    chain.Database,
	}

	return iter
}

func (iter *Iterator) Next() *Block {
	var block *Block

	encodedBlock, err := iter.Database.Get(iter.CurrentHash, nil)
	Handle(err)

	block = Deserialize(encodedBlock)

	iter.CurrentHash = block.PreviousHash

	return block
}
