package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type EthEventLog struct {
	Blockchain       *string  `json:"blockchain"`
	Address          *string  `json:"address"`
	BlockHash        *string  `json:"blockHash"`
	BlockNumber      *string  `json:"blockNumber"`
	BlockTimestamp   *string  `json:"blockTimestamp"`
	Data             *string  `json:"data"`
	LogIndex         *string  `json:"logIndex"`
	Removed          *bool    `json:"removed"`
	Topics           []string `json:"topics"`
	TransactionHash  *string  `json:"transactionHash"`
	TransactionIndex *string  `json:"transactionIndex"`
}

type EthTransaction struct {
	BlockHash        *string `json:"blockHash"`
	BlockNumber      *string `json:"blockNumber"`
	ChainID          *string `json:"chainId"`
	From             *string `json:"from"`
	Gas              *string `json:"gas"`
	GasPrice         *string `json:"gasPrice"`
	Hash             *string `json:"hash"`
	Input            *string `json:"input"`
	Nonce            *string `json:"nonce"`
	R                *string `json:"r"`
	S                *string `json:"s"`
	To               *string `json:"to"`
	TransactionIndex *string `json:"transactionIndex"`
	Type             *string `json:"type"`
	V                *string `json:"v"`
	Value            *string `json:"value"`
}

type EthTransactionReceipt struct {
	BlockHash         *string       `json:"blockHash"`
	BlockNumber       *string       `json:"blockNumber"`
	ContractAddress   *string       `json:"contractAddress"`
	CumulativeGasUsed *string       `json:"cumulativeGasUsed"`
	EffectiveGasPrice *string       `json:"effectiveGasPrice"`
	From              *string       `json:"from"`
	GasUsed           *string       `json:"gasUsed"`
	Logs              []EthEventLog `json:"logs"`
	LogsBloom         *string       `json:"logsBloom"`
	Root              *string       `json:"root"`
	Status            *string       `json:"status"`
	To                *string       `json:"to"`
	TransactionHash   *string       `json:"transactionHash"`
	TransactionIndex  *string       `json:"transactionIndex"`
	Type              *string       `json:"type"`
}

type EthResponse struct {
	JsonRpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Error   any    `json:"error"`
	Result  any    `json:"result"`
}

type EthBlockRangeError struct {
	Code    float64                `json:"code"`
	Message string                 `json:"message"`
	Data    EthBlockRangeErrorData `json:"data"`
}

type EthBlockRangeErrorData struct {
	From  string  `json:"from"`
	Limit float64 `json:"limit"`
	To    string  `json:"to"`
}

type EthErrorHandler struct {
	Err           error
	BlockInterval []uint64
}

const (
	InfuraBaseUrl = "https://mainnet.infura.io/v3/"
)

func parseEthResponse(response *EthResponse, output interface{}) (errHandler *EthErrorHandler) {
	errHandler = &EthErrorHandler{
		Err:           nil,
		BlockInterval: []uint64{},
	}
	if response.Error != nil {
		message := "an error occurred on fetching data from Infura API !"
		if reflect.TypeOf(response.Error).Kind() == reflect.Map {
			code := response.Error.(map[string]interface{})["code"].(float64)
			message = response.Error.(map[string]interface{})["message"].(string)
			message = fmt.Sprintf("<%f> %s", code, message)
			if code == -32005 && strings.Contains(message, "query returned more than 10000 results") {
				errorStr, err := json.Marshal(response.Error)
				if err != nil {
					errHandler.Err = err
					return
				}
				errorPayload := &EthBlockRangeError{}
				err = json.Unmarshal(errorStr, errorPayload)
				if err != nil {
					errHandler.Err = err
					return
				}
				bnFrom, err := hexutil.DecodeUint64(errorPayload.Data.From)
				if err != nil {
					errHandler.Err = err
					return
				}
				bnTo, err := hexutil.DecodeUint64(errorPayload.Data.To)
				if err != nil {
					errHandler.Err = err
					return
				}
				errHandler.BlockInterval = []uint64{bnFrom, bnTo}
				return
			}
		}
		errHandler.Err = errors.New(message)
		return
	}
	resJson, err := json.Marshal(response.Result)
	if err != nil {
		errHandler.Err = err
		return
	}
	err = json.Unmarshal(resJson, output)
	if err != nil {
		errHandler.Err = err
	}
	return
}

func InfuraRequest(payload map[string]any, output interface{}) (errHandler *EthErrorHandler) {
	errHandler = &EthErrorHandler{
		Err:           nil,
		BlockInterval: []uint64{},
	}

	rv := reflect.ValueOf(output)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		err := errors.New("invalid result handler. Must be a pointer to a non nil value")
		errHandler.Err = err
		return
	}

	url := InfuraBaseUrl + os.Getenv("INFURA_API_KEY")

	response := &EthResponse{}
	err := SendHttpRequest(url, "POST", nil, payload, response)
	if err != nil {
		errHandler.Err = err
		return
	}

	errHandler = parseEthResponse(response, output)

	return
}
