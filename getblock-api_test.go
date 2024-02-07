package main

import (
	"context"
	"encoding/hex"
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
	lastBlockNumHex, err := api.GetBlockNumber()

	if err != nil {
		t.Errorf("GetBlockNumber() failed: %s", err)
	}

	fmt.Printf("Block number: %s\n", lastBlockNumHex)
	// assert the prevblockNumber contains 0x in the beginning
	assert.Contains(t, lastBlockNumHex, "0x")

	prevBlockNum, err := api.GetEthBlockPrevNumber(lastBlockNumHex)
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

	api := NewGetBlockAPI()
	ctx := context.Background()
	blocks, err := api.GetNLastBlocks(ctx, 2)
	if err != nil {
		t.Errorf("GetNLastBlocks() failed: %s", err)
	}

	if len(blocks) != 5 {
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
	contractCodeHex, err := api.GetContractCode(ctx, contractAddress)
	if err != nil {
		t.Errorf("GetContractCode() failed: %s", err)
	}

	fmt.Printf("Contract code: %s\n", contractCodeHex)

	_, err = hex.DecodeString(contractCodeHex[2:])
	if err != nil {
		t.Errorf("hex.DecodeString() failed: %s", err)
	}

	// convert code to the string and assert it's not empty

	assert.NotEmpty(t, contractCodeHex)
}
