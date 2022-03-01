package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/wilmacedo/willchain-go/core"
	"github.com/wilmacedo/willchain-go/factory"
	"github.com/wilmacedo/willchain-go/models"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" balance -address [ADDRESS] - Get the balance of address")
	fmt.Println(" createblockchain -address [ADDRESS] - Creates a blockchain in another address")
	fmt.Println(" printchain - Prints the blocks in the chain")
	fmt.Println(" send -from [FROM] -to [TO] -amount [AMOUNT] - Send amount from to another account and specificy amount")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) printChain() {
	chain := factory.ContinueBlockchain("")
	defer chain.Database.Close()

	iter := chain.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("Previous hash: %x\n", block.PreviousHash)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := factory.NewProof(block)
		fmt.Printf("Is valide: %s\n\n", strconv.FormatBool(pow.Validate()))

		if len(block.PreviousHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createBlockchain(address string) {
	chain := factory.InitBlockchain(address)
	defer chain.Database.Close()

	fmt.Println("Finished!")
}

func (cli *CommandLine) getBalance(address string) {
	chain := factory.ContinueBlockchain(address)
	defer chain.Database.Close()

	balance := 0
	txs := chain.FindResTX(address)

	for _, tx := range txs {
		balance += tx.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int) {
	chain := factory.ContinueBlockchain(from)
	defer chain.Database.Close()

	tx := factory.NewTransaction(from, to, amount, chain)
	chain.AddBlock([]*models.Transaction{tx})

	fmt.Println("Success!")
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	balanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	balanceAddress := balanceCmd.String("address", "", "The address to retrieve balance")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to be create")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "balance":
		err := balanceCmd.Parse(os.Args[2:])
		core.Handle(err)

	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		core.Handle(err)

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		core.Handle(err)

	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		core.Handle(err)

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if balanceCmd.Parsed() {
		if *balanceAddress == "" {
			balanceCmd.Usage()
			runtime.Goexit()
		}

		cli.getBalance(*balanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}

		cli.createBlockchain(*createBlockchainAddress)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" {
			sendCmd.Usage()
			runtime.Goexit()
		}

		if *sendTo == "" {
			sendCmd.Usage()
			runtime.Goexit()
		}

		if *sendAmount < 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}
