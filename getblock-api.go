package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	GetBlockEndpoint   = "https://go.getblock.io/"
	GetBlockAPITimeout = 5 * time.Second
)

type GetBlockPostRequestModel struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	Id      string `json:"id"`
}

type GetBlockGenericResult[T any] struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  T           `json:"result"`
	Id      string      `json:"id"`
	Error   interface{} `json:"error,omitempty"`
}

type EthereumBlock struct {
	Difficulty       string        `json:"difficulty"`
	ExtraData        string        `json:"extraData"`
	GasLimit         string        `json:"gasLimit"`
	GasUsed          string        `json:"gasUsed"`
	Hash             string        `json:"hash"`
	LogsBloom        string        `json:"logsBloom"`
	Miner            string        `json:"miner"`
	MixHash          string        `json:"mixHash"`
	Nonce            string        `json:"nonce"`
	Number           string        `json:"number"`
	ParentHash       string        `json:"parentHash"`
	ReceiptsRoot     string        `json:"receiptsRoot"`
	Sha3Uncles       string        `json:"sha3Uncles"`
	Size             string        `json:"size"`
	StateRoot        string        `json:"stateRoot"`
	Timestamp        string        `json:"timestamp"`
	TotalDifficulty  string        `json:"totalDifficulty"`
	Transactions     []Transaction `json:"transactions"`
	TransactionsRoot string        `json:"transactionsRoot"`
	Uncles           []interface{} `json:"uncles"`
}

type Transaction struct {
	AccessList           []interface{} `json:"accessList"`
	BlockHash            string        `json:"blockHash"`
	BlockNumber          string        `json:"blockNumber"`
	ChainID              string        `json:"chainId"`
	From                 string        `json:"from"`
	Gas                  string        `json:"gas"`
	GasPrice             string        `json:"gasPrice"`
	Hash                 string        `json:"hash"`
	Input                string        `json:"input"`
	MaxFeePerGas         string        `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string        `json:"maxPriorityFeePerGas"`
	Nonce                string        `json:"nonce"`
	R                    string        `json:"r"`
	S                    string        `json:"s"`
	To                   string        `json:"to"`
	TransactionIndex     string        `json:"transactionIndex"`
	Type                 string        `json:"type"`
	V                    string        `json:"v"`
	Value                string        `json:"value"`
}

type GetBlockAPI struct {
	apiTimeout time.Duration
}

func NewGetBlockAPI() *GetBlockAPI {
	return &GetBlockAPI{
		apiTimeout: GetBlockAPITimeout * time.Second,
	}
}

func (api *GetBlockAPI) GetBlockNumber(ctx context.Context) (string, error) {
	res := &GetBlockGenericResult[string]{}
	err := api.Call(ctx, "eth_blockNumber", nil, res)
	if err != nil {
		return "", err
	}

	if res.Error != nil {
		return "", fmt.Errorf("eth_blockNumber failed: %s", res.Error)
	}

	return res.Result, nil
}

func (api *GetBlockAPI) GetBlockByNumber(ctx context.Context, n string) (*EthereumBlock, error) {
	res := &GetBlockGenericResult[EthereumBlock]{}
	err := api.Call(ctx, "eth_getBlockByNumber", []any{n, true}, res)
	if err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber failed: %s", err)
	}

	if res.Error != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber failed: %s", res.Error)
	}

	return &res.Result, nil
}

func (api *GetBlockAPI) GetNLastBlocks(ctx context.Context, n int) ([]EthereumBlock, error) {
	var blocks []EthereumBlock
	lastBlockNumber, err := api.GetBlockNumber(ctx)
	if err != nil {
		return nil, err
	}

	latestBlock, err := api.GetBlockByNumber(ctx, lastBlockNumber)
	if err != nil {
		return nil, err
	}
	blocks = append(blocks, *latestBlock)
	prevBlockNumber, err := api.WhichPrevBlockNumber(latestBlock.Number)
	if err != nil {
		return nil, err
	}

	for i := 0; i < n-1; i++ {
		block, err := api.GetBlockByNumber(ctx, prevBlockNumber)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, *block)

		prevBlockNumber, err = api.WhichPrevBlockNumber(block.Number)
		if err != nil {
			return nil, err
		}
	}

	fmt.Printf("Fetched %d blocks\n", len(blocks))
	return blocks, nil
}

// Returns the code of the smart contract at the specified address.
func (api *GetBlockAPI) GetContractCode(ctx context.Context, address string) (*Contract, error) {
	res := &GetBlockGenericResult[string]{}

	err := api.Call(ctx, "eth_getCode", []any{address, "latest"}, res)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart contract code: %s", err)
	}

	if res.Error != nil {
		return nil, fmt.Errorf("failed to get smart contract code: %s", res.Error)
	}

	return &Contract{
		Address:  address,
		ByteCode: res.Result,
	}, nil
}

// Returns the number of the block before the specified one.
func (api *GetBlockAPI) WhichPrevBlockNumber(blockNumHex string) (string, error) {
	decimalBlockNumber, err := strconv.ParseInt(blockNumHex[2:], 16, 64)
	if err != nil {
		return "", fmt.Errorf("failed to parse block number: %s", err)
	}

	prevBlockNumber := decimalBlockNumber - 1
	prevBlockNumberHex := strconv.FormatInt(prevBlockNumber, 16)
	return fmt.Sprintf("0x%s", prevBlockNumberHex), nil

}

type CallResult struct {
	Result interface{}
	Error  error
}

func (api *GetBlockAPI) Call(ctx context.Context, method string, params []any, res interface{}) error {
	payload := &GetBlockPostRequestModel{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		Id:      "getblock.io",
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	resch := make(chan CallResult)
	defer close(resch)

	go func() {
		paramsJson, err := json.Marshal(payload)
		if err != nil {
			resch <- CallResult{Error: fmt.Errorf("failed to marshal payload: %s", err)}
			return
		}

		d := bytes.NewBuffer(paramsJson)
		endpoint := getEndpointWithApiKey()

		resp, err := http.Post(endpoint, "application/json", d)
		if err != nil {
			resch <- CallResult{Error: fmt.Errorf("failed to post request: %s", err)}
			return
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			resch <- CallResult{Error: fmt.Errorf("failed to read response: %s, err: %s ", string(data), err)}
			return
		}
		if resp.StatusCode != http.StatusOK {
			resch <- CallResult{Error: fmt.Errorf("failed request with status=%d, resp: %v", resp.StatusCode, string(data))}
			return
		}

		if err = json.Unmarshal(data, res); err != nil {
			resch <- CallResult{Error: fmt.Errorf("failed to unmarshal response: %#v, err: %s", string(data), err)}
			return
		}

		resch <- CallResult{Result: res}
	}()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout fron getbkoock.io API call")
		case r := <-resch:
			res = r.Result
			return r.Error
		}
	}
}

func getEndpointWithApiKey() string {
	GETBLOCK_API_KEY := os.Getenv("GETBLOCK_API_KEY")
	if GETBLOCK_API_KEY == "" {
		panic("GETBLOCK_API_KEY is not set")
	}

	return fmt.Sprintf("%s%s", GetBlockEndpoint, GETBLOCK_API_KEY)
}
