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
)

const Endpoint = "https://go.getblock.io/"

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
}

func NewGetBlockAPI() *GetBlockAPI {
	return &GetBlockAPI{}
}

func (api *GetBlockAPI) GetBlockNumber() (string, error) {
	res := &GetBlockGenericResult[string]{}
	err := api.Call("eth_blockNumber", nil, res)
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
	err := api.Call("eth_getBlockByNumber", []any{n, true}, res)
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
	lastBlockNumber, err := api.GetBlockNumber()
	if err != nil {
		return nil, fmt.Errorf("eth_blockNumber failed: %s", err)
	}

	latestBlock, err := api.GetBlockByNumber(ctx, lastBlockNumber)
	if err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber failed: %s", err)
	}
	blocks = append(blocks, *latestBlock)
	prevBlockNumber, err := api.GetEthBlockPrevNumber(latestBlock.Number)
	if err != nil {
		return nil, fmt.Errorf("GetEthBlockPrevNumber failed: %s", err)
	}

	for i := 0; i < n-1; i++ {
		block, err := api.GetBlockByNumber(ctx, prevBlockNumber)
		if err != nil {
			return nil, fmt.Errorf("eth_getBlockByNumber failed: %s", err)
		}
		blocks = append(blocks, *block)

		prevBlockNumber, err = api.GetEthBlockPrevNumber(block.Number)
		if err != nil {
			return nil, fmt.Errorf("GetEthBlockPrevNumber failed: %s", err)
		}
	}

	fmt.Printf("Fetched %d blocks\n", len(blocks))
	return blocks, nil
}

// Returns the code of the smart contract at the specified address. Besustores compiled smart contract code as a hexadecimal value.
func (api *GetBlockAPI) GetContractCode(ctx context.Context, address string) (string, error) {
	res := &GetBlockGenericResult[string]{}

	err := api.Call("eth_getCode", []any{address, "latest"}, res)
	if err != nil {
		return "", fmt.Errorf("eth_getCode failed: %s", err)
	}

	if res.Error != nil {
		return "", fmt.Errorf("eth_getCode failed: %s", res.Error)
	}

	return res.Result, nil
}

func (api *GetBlockAPI) GetEthBlockPrevNumber(blockNumHex string) (string, error) {
	decimalBlockNumber, err := strconv.ParseInt(blockNumHex[2:], 16, 64)
	if err != nil {
		return "", fmt.Errorf("failed to parse block number: %s", err)
	}

	prevBlockNumber := decimalBlockNumber - 1
	prevBlockNumberHex := strconv.FormatInt(prevBlockNumber, 16)
	return fmt.Sprintf("0x%s", prevBlockNumberHex), nil

}

func (api *GetBlockAPI) Call(method string, params []any, res interface{}) error {
	payload := &GetBlockPostRequestModel{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		Id:      "getblock.io",
	}

	paramsJson, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	d := bytes.NewBuffer(paramsJson)
	endpoint := getEndpoint()

	resp, err := http.Post(endpoint, "application/json", d)
	if err != nil {
		return fmt.Errorf("failed to make POST request: %s", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %s", err)
	}

	if err = json.Unmarshal(data, res); err != nil {
		return fmt.Errorf("failed to unmarshal response: %s", err)
	}

	return nil
}

func getEndpoint() string {
	GETBLOCK_API_KEY := os.Getenv("GETBLOCK_API_KEY")
	if GETBLOCK_API_KEY == "" {
		panic("GETBLOCK_API_KEY is not set")
	}

	return fmt.Sprintf("%s%s", Endpoint, GETBLOCK_API_KEY)
}
