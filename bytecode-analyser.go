package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wealdtech/go-merkletree/keccak256"
)

// The bytecode of a contract contains instructions that compare the first
// four bytes of the call data to the signatures of its functions
type ByteCodeAnalyser struct {
	getBlockAPI *GetBlockAPI
}

func NewByteCodeAnalyser(getBlockAPI *GetBlockAPI) *ByteCodeAnalyser {
	return &ByteCodeAnalyser{
		getBlockAPI: getBlockAPI,
	}
}

func (b *ByteCodeAnalyser) contractMetodSignature(method string, params []interface{}) string {
	a := keccak256.New()

	str := method + "("
	for i, param := range params {
		str += param.(string)
		if i < len(params)-1 {
			str += ","
		}
	}
	str += ")"

	hash := a.Hash([]byte(str))
	result := fmt.Sprintf("%x", hash)

	return string(result[:8])
}

func (b *ByteCodeAnalyser) SatisfiesERC20(bytecode string) error {
	totalSupply := b.contractMetodSignature("totalSupply", []interface{}{})
	balanceOf := b.contractMetodSignature("balanceOf", []interface{}{"address"})
	transfer := b.contractMetodSignature("transfer", []interface{}{"address", "uint256"})
	transferFrom := b.contractMetodSignature("transferFrom", []interface{}{"address", "address", "uint256"})
	approve := b.contractMetodSignature("approve", []interface{}{"address", "uint256"})
	allowance := b.contractMetodSignature("allowance", []interface{}{"address", "address"})

	transferEvent := b.contractMetodSignature("Transfer", []interface{}{"address", "address", "uint256"})
	approvalEvent := b.contractMetodSignature("Approval", []interface{}{"address", "address", "uint256"})

	for i, sig := range map[string]string{
		"totalSupply()":                         totalSupply,
		"balanceOf(address)":                    balanceOf,
		"transfer(address,uint256)":             transfer,
		"transferFrom(address,address,uint256)": transferFrom,
		"approve(address,uint256)":              approve,
		"allowance(address,address)":            allowance,
		"Transfer(address,address,uint256)":     transferEvent,
		"Approval(address,address,uint256)":     approvalEvent,
	} {
		if !strings.Contains(bytecode, sig) {
			return fmt.Errorf("signature %s not found in bytecode %v", sig, i)
		}
	}

	return nil
}

func (b *ByteCodeAnalyser) FilterOnlyERC20CompatibleTx(txList []Transaction) []Transaction {
	var result []Transaction

	for _, tx := range txList {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		contract, err := b.getBlockAPI.GetContractCode(ctx, tx.To)
		if err != nil {
			fmt.Println("GetContractCode() failed: \n" + err.Error())
			continue
		}

		if len(contract) == 0 {
			fmt.Println("contract is empty")
			continue
		}

		if err := b.SatisfiesERC20(contract); err != nil {
			// fmt.Printf("contract %s does not satisfy ERC20: %s\n", tx.To, err.Error())
			continue
		}

		// fmt.Println("compatible  contract: " + tx.To)
		result = append(result, tx)
	}

	return result
}
