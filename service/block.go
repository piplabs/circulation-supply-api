package service

import (
	"bytes"
	"circulation-supply-api/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
)

type BlockResponse struct {
	Result Block `json:"result"`
}

type Block struct {
	Number        string       `json:"number"`
	Time          string       `json:"timestamp"`
	BaseFeePerGas string       `json:"baseFeePerGas"`
	GasUsed       string       `json:"gasUsed"`
	Withdrawals   []Withdrawal `json:"withdrawals"`
}

type Withdrawal struct {
	Amount    string `json:"amount"`
	Validator string `json:"validatorIndex"`
	Address   string `json:"address"`
}

func getBlockByNumber(blockNumber string) (*Block, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{blockNumber, true},
		"id":      1,
	}

	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(config.Conf.RpcEndpoint, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var blockResp BlockResponse
	if err := json.Unmarshal(body, &blockResp); err != nil {
		return nil, err
	}

	return &blockResp.Result, nil
}

func getBalanceAtBlock(address string, blockNumber uint64) (*big.Int, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBalance",
		"params":  []interface{}{address, fmt.Sprintf("0x%x", blockNumber)},
		"id":      1,
	}
	response, err := rpcCall(payload)
	if err != nil {
		return nil, err
	}

	result, ok := response["result"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to assert result as string")
	}

	balance, success := new(big.Int).SetString(result[2:], 16)
	if !success {
		return nil, fmt.Errorf("invalid balance format")
	}
	return balance, nil
}

func rpcCall(payload map[string]interface{}) (map[string]interface{}, error) {
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := http.Post(config.Conf.RpcEndpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send RPC request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if err, exists := response["error"]; exists {
		return nil, fmt.Errorf("RPC error: %v", err)
	}

	return response, nil
}
