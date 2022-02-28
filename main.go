package main

import (
	"fmt"

	"github.com/wilmacedo/willchain-go/models"
)

func main() {
	chain := models.InitBlockchain()

	chain.AddBlock("First block after genesis")
	chain.AddBlock("Second block after genesis")
	chain.AddBlock("Third block after genesis")

	for _, block := range chain.Blocks {
		fmt.Printf("Block hash: %x\n", block.Hash)
		fmt.Printf("Block data: %s\n\n", block.Data)
	}
}
