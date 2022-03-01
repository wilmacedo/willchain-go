package models

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"

	"github.com/wilmacedo/willchain-go/core"
)

type Transaction struct {
	Hash     []byte
	Requests []TXRequest
	Results  []TXResult
}

type TXRequest struct {
	Hash []byte
	Out  int
	Sig  string
}

type TXResult struct {
	Value  int
	PubKey string
}

func (tx *Transaction) CalculateHash() {
	var encoded bytes.Buffer
	var hash [32]byte

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)
	core.Handle(err)

	hash = sha256.Sum256(encoded.Bytes())
	tx.Hash = hash[:]
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Requests) == 1 && len(tx.Requests[0].Hash) == 0 && tx.Requests[0].Out == -1
}

func (req *TXRequest) CanUnlock(data string) bool {
	return req.Sig == data
}

func (res *TXResult) CanBeUnlocked(data string) bool {
	return res.PubKey == data
}
