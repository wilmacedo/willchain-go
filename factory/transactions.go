package factory

import (
	"encoding/hex"
	"fmt"

	"github.com/wilmacedo/willchain-go/core"
	"github.com/wilmacedo/willchain-go/models"
)

func CoinbaseTX(to, data string) *models.Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txReq := models.TXRequest{
		Hash: []byte{},
		Out:  -1,
		Sig:  data,
	}

	txResp := models.TXResult{
		Value:  100,
		PubKey: to,
	}

	tx := &models.Transaction{
		Hash:     nil,
		Requests: []models.TXRequest{txReq},
		Results:  []models.TXResult{txResp},
	}

	tx.CalculateHash()

	return tx
}

func NewTransaction(from, to string, amount int, chain *Blockchain) *models.Transaction {
	var requests []models.TXRequest
	var results []models.TXResult

	acc, validResults := chain.FindSpendableResults(from, amount)

	if acc < amount {
		core.Handle(core.ErrEnoughFunds)
	}

	for txhash, res := range validResults {
		txHash, err := hex.DecodeString(txhash)
		core.Handle(err)

		for _, rs := range res {
			request := models.TXRequest{
				Hash: txHash,
				Out:  rs,
				Sig:  from,
			}
			requests = append(requests, request)
		}
	}

	results = append(results, models.TXResult{
		Value:  amount,
		PubKey: to,
	})

	if acc > amount {
		results = append(results, models.TXResult{
			Value:  acc - amount,
			PubKey: from,
		})
	}

	tx := &models.Transaction{
		Hash:     nil,
		Requests: requests,
		Results:  results,
	}
	tx.CalculateHash()

	return tx
}
