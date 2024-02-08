package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/wealdtech/go-merkletree/keccak256"
)

var batchSize = 2

// The bytecode of a contract contains instructions that compare the first
// four bytes of the call data to the signatures of its functions
type ByteCodeAnalyser struct {
	getBlockAPI *GetBlockAPI
}

type Contract struct {
	Address  string
	ByteCode string
}

func (c *Contract) IsContainAllSignatures(signatures map[string]string) bool {
	for _, signature := range signatures {
		if !strings.Contains(c.ByteCode, signature) {
			return false
		}
	}

	return true

}

func NewByteCodeAnalyser(getBlockAPI *GetBlockAPI) *ByteCodeAnalyser {
	return &ByteCodeAnalyser{
		getBlockAPI: getBlockAPI,
	}
}

func (b *ByteCodeAnalyser) FilterOnlyERC20CompatibleTx(
	ctx context.Context, txList []Transaction,
) []Transaction {
	result := make([]Transaction, 0)
	var resultch = make(chan Transaction)

	if len(txList) == 0 {
		return result
	}

	erc20Signatures := b.calcERC20Signatures()
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		b.batch(ctx, txList, erc20Signatures, resultch)
		wg.Done()
	}()

	fmt.Println("Waiting for the result")

	for tx := range resultch {
		result = append(result, tx)
	}

	fmt.Printf("Found %d ERC20 compatible transactions\n", len(result))
	wg.Wait()
	return result
}

func (b *ByteCodeAnalyser) batch(
	ctx context.Context,
	txList []Transaction,
	signaturesToInclude map[string]string,
	resultCh chan Transaction,
) {
	ch := make(chan struct{}, batchSize)
	defer close(ch)
	defer close(resultCh)
	var wg sync.WaitGroup

	for _, tx := range txList {
		wg.Add(1)
		ch <- struct{}{}
		go func(tx Transaction) {
			defer func() {
				wg.Done()
				<-ch
			}()

			contract, err := b.getBlockAPI.GetContractCode(ctx, tx.To)
			if err != nil {
				fmt.Println("GetContractCode failed: \n" + err.Error())
				return
			}

			if len(contract.ByteCode) == 0 {
				fmt.Println("contract is empty")
				return
			}

			if contract.IsContainAllSignatures(signaturesToInclude) {
				resultCh <- tx
				fmt.Printf("found contract %s is ERC20 compatible\n", tx.Hash)
			}
		}(tx)
	}

	wg.Wait()
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

func (b *ByteCodeAnalyser) calcERC20Signatures() map[string]string {
	totalSupply := b.contractMetodSignature("totalSupply", []interface{}{})
	balanceOf := b.contractMetodSignature("balanceOf", []interface{}{"address"})
	transfer := b.contractMetodSignature("transfer", []interface{}{"address", "uint256"})
	transferFrom := b.contractMetodSignature("transferFrom", []interface{}{"address", "address", "uint256"})
	approve := b.contractMetodSignature("approve", []interface{}{"address", "uint256"})
	allowance := b.contractMetodSignature("allowance", []interface{}{"address", "address"})

	transferEvent := b.contractMetodSignature("Transfer", []interface{}{"address", "address", "uint256"})
	approvalEvent := b.contractMetodSignature("Approval", []interface{}{"address", "address", "uint256"})

	return map[string]string{
		"totalSupply()":                         totalSupply,
		"balanceOf(address)":                    balanceOf,
		"transfer(address,uint256)":             transfer,
		"transferFrom(address,address,uint256)": transferFrom,
		"approve(address,uint256)":              approve,
		"allowance(address,address)":            allowance,
		"Transfer(address,address,uint256)":     transferEvent,
		"Approval(address,address,uint256)":     approvalEvent,
	}
}
