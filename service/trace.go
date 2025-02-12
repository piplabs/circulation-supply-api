package service

import (
	"bytes"
	"circulation-supply-api/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"

	log "log/slog"
)

type InternalTx struct {
	From    string       `json:"from"`
	Gas     string       `json:"gas"`
	GasUsed string       `json:"gasUsed"`
	To      string       `json:"to"`
	Input   string       `json:"input"`
	Output  string       `json:"output"`
	Value   string       `json:"value"`
	Type    string       `json:"type"`
	Calls   []InternalTx `json:"calls"`
}

type ParentTx struct {
	TxHash string `json:"txHash"`
	Result struct {
		From    string       `json:"from"`
		Gas     string       `json:"gas"`
		GasUsed string       `json:"gasUsed"`
		To      string       `json:"to"`
		Input   string       `json:"input"`
		Calls   []InternalTx `json:"calls"` // Internal transactions are here
		Value   string       `json:"value"`
		Type    string       `json:"type"`
	} `json:"result"`
}

type Trace struct {
	BlockNumber string
	ParentTxs   []ParentTx
}

func GetStakedToken(blockNumber string) (*big.Float, error) {
	// Fetch internal transactions
	trace, err := FetchTraces(blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch traces: %v", err)
	}

	// Process transactions
	stakedToken := ProcessTraces(trace)
	return stakedToken, nil
}

func FetchTraces(blockNumber string) (*Trace, error) {
	requestBody := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "debug_traceBlockByNumber",
		"params": []interface{}{
			blockNumber,
			map[string]string{"tracer": "callTracer"},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(config.Conf.RpcEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send RPC request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var result struct {
		Result []ParentTx `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	return &Trace{
		BlockNumber: blockNumber,
		ParentTxs:   result.Result,
	}, nil
}

func ProcessTraces(trace *Trace) *big.Float {
	transactions := trace.ParentTxs

	threshold := big.NewFloat(config.Conf.StakeThreshold) // 1024 IP

	stakedToken := big.NewFloat(0)
	internalTxList := make([]InternalTx, 0, 16)

	for _, tx := range transactions {
		internalTxList = append(internalTxList, tx.Result.Calls...)
	}

	for len(internalTxList) > 0 {
		internalTx := internalTxList[0]
		internalTxList = internalTxList[1:]
		if len(internalTx.Calls) > 0 {
			internalTxList = append(internalTxList, internalTx.Calls...)
		}
		if !strings.EqualFold(internalTx.From, config.Conf.StakeContractAddress) ||
			!strings.EqualFold(internalTx.To, config.Conf.StakeReceiverAddress) {
			continue
		}

		valueWei := new(big.Int)
		valueWei.SetString(internalTx.Value[2:], 16)

		valueEth := new(big.Float).SetInt(valueWei)
		ethValue := new(big.Float).Quo(valueEth, big.NewFloat(1e18))

		if ethValue.Cmp(threshold) >= 0 {
			log.Info("High Value Internal Transaction (≥ 1024 ETH)", "Value", ethValue.Text('f', 6), "blockNumber", trace.BlockNumber)
			stakedToken.Add(stakedToken, ethValue)
		}
	}
	return stakedToken
}
