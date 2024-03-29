package main

import (
	"context"
	"fmt"

	"github.com/joho/godotenv"
)

var DebubMode = true

func main() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	getBlockAPI := NewGetBlockAPI()
	analyser := NewByteCodeAnalyser(getBlockAPI)
	activityCounter := NewActivityCounter()

	ctx := context.Background()

	blocks, err := getBlockAPI.GetNLastBlocks(ctx, 1)
	if err != nil {
		panic(err)
	}

	for i, block := range blocks {
		erc20CompatibleTx := analyser.FilterOnlyERC20CompatibleTx(ctx, block.Transactions)

		for _, tx := range erc20CompatibleTx {
			activityCounter.AddTx(tx)
		}
		fmt.Printf("Block %d: %d transactions, %d ERC20 compatible\n", i, len(block.Transactions), len(erc20CompatibleTx))
	}

	result := activityCounter.GetMostMostActive(5)
	for _, r := range result {
		fmt.Printf("Address: %s, ActivityCount: %d\n", r.Address, r.ActivityCount)
	}
}
