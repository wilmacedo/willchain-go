package models

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"

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

func CoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txReq := TXRequest{
		Hash: []byte{},
		Out:  -1,
		Sig:  data,
	}

	txResp := TXResult{
		Value:  100,
		PubKey: to,
	}

	tx := &Transaction{
		Hash:     nil,
		Requests: []TXRequest{txReq},
		Results:  []TXResult{txResp},
	}

	tx.CalculateHash()

	return tx
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

func NewTransaction(from, to string, amount int, chain *Blockchain) *Transaction {
	var requests []TXRequest
	var results []TXResult

	acc, validResults := chain.FindSpendableResults(from, amount)

	if acc < amount {
		core.Handle(core.ErrEnoughFunds)
	}

	for txhash, res := range validResults {
		txHash, err := hex.DecodeString(txhash)
		core.Handle(err)

		for _, rs := range res {
			request := TXRequest{
				Hash: txHash,
				Out:  rs,
				Sig:  from,
			}
			requests = append(requests, request)
		}
	}

	results = append(results, TXResult{
		Value:  amount,
		PubKey: to,
	})

	if acc > amount {
		results = append(results, TXResult{
			Value:  acc - amount,
			PubKey: from,
		})
	}

	tx := &Transaction{
		Hash:     nil,
		Requests: requests,
		Results:  results,
	}
	tx.CalculateHash()

	return tx
}
