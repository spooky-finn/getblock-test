package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestGetPrevBlockNumber(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		t.Errorf("Error loading .env file: %s", err)
	}

	api := NewGetBlockAPI()
	lastBlockNumHex, err := api.GetBlockNumber(context.Background())

	if err != nil {
		t.Errorf("GetBlockNumber() failed: %s", err)
	}

	fmt.Printf("Block number: %s\n", lastBlockNumHex)
	// assert the prevblockNumber contains 0x in the beginning
	assert.Contains(t, lastBlockNumHex, "0x")

	prevBlockNum, err := api.WhichPrevBlockNumber(lastBlockNumHex)
	if err != nil {
		t.Errorf("GetEthBlockPrevNumber() failed: %s", err)
	}

	fmt.Printf("Previous block number: %s\n", prevBlockNum)
	// assert the prevblockNumber contains 0x in the beginning
	assert.Contains(t, prevBlockNum, "0x")
}

func TestGetNLastBlocks(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		t.Errorf("Error loading .env file: %s", err)
	}
	N := 2
	api := NewGetBlockAPI()
	ctx := context.Background()
	blocks, err := api.GetNLastBlocks(ctx, N)
	if err != nil {
		t.Errorf("GetNLastBlocks() failed: %s", err)
	}

	if len(blocks) != N {
		t.Errorf("GetNLastBlocks() returned %d blocks, expected 5", len(blocks))
	}

	// each block number should be unique and contain 0x in the beginning
	seen := make(map[string]bool)
	for _, block := range blocks {
		assert.Contains(t, block.Number, "0x")
		_, ok := seen[block.Number]
		assert.False(t, ok)
		seen[block.Number] = true
	}
}

func TestGetContractCode(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		t.Errorf("Error loading .env file: %s", err)
	}

	api := NewGetBlockAPI()
	contractAddress := "0x06012c8cf97bead5deae237070f9587f8e7a266d"
	ctx := context.Background()
	contract, err := api.GetContractCode(ctx, contractAddress)
	if err != nil {
		t.Errorf("GetContractCode() failed: %s", err)
	}

	fmt.Printf("Contract code: %s\n", contract.ByteCode)
	assert.NotEmpty(t, contractAddress)
	assert.NotEmpty(t, contract.ByteCode)
}
