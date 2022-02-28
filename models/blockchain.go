package models

type Blockchain struct {
	Blocks []*Block
}

func (chain *Blockchain) AddBlock(data string) {
	previousBlock := chain.Blocks[len(chain.Blocks)-1]
	new := CreateBlock(data, previousBlock.Hash)

	chain.Blocks = append(chain.Blocks, new)
}

func InitBlockchain() *Blockchain {
	chain := &Blockchain{
		Blocks: []*Block{Genesis()},
	}

	return chain
}
