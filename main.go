package main

import (
	"os"

	"github.com/wilmacedo/willchain-go/cli"
	"github.com/wilmacedo/willchain-go/models"
)

func main() {
	defer os.Exit(0)

	chain := models.InitBlockchain()
	defer chain.Database.Close()

	cli := cli.CommandLine{
		Blockchain: chain,
	}
	cli.Run()
}
