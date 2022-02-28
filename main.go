package main

import (
	"fmt"
	"strconv"

	"github.com/wilmacedo/willchain-go/models"
)

func main() {
	chain := models.InitBlockchain()

	chain.AddBlock("Second")
	chain.AddBlock("Third")
	chain.AddBlock("Fourth")

	for _, block := range chain.Blocks {
		fmt.Printf("Previous hash: %x\n", block.PreviousHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Data: %s\n", block.Data)

		pow := models.NewProof(block)
		fmt.Printf("Is valide: %s\n\n", strconv.FormatBool(pow.Validate()))
	}
}
