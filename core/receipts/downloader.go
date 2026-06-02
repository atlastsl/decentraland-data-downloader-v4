package receipts

import (
	"context"
	"decentraland-data-downloader-v4/core/hashes"
	"decentraland-data-downloader-v4/packages/api"
	"decentraland-data-downloader-v4/packages/database"
	"decentraland-data-downloader-v4/packages/helpers"
	"errors"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	downloadAttempts = 10
)

func getTransactionByHash(transactionHash string) (txInfo *api.EthTransaction, err error) {
	payload := map[string]any{
		"jsonrpc": "2.0",
		"method":  "eth_getTransactionByHash",
		"id":      time.Now().UnixMilli(),
		"params":  []string{transactionHash},
	}
	for i := 0; i < downloadAttempts; i++ {
		txInfo = &api.EthTransaction{}
		errHandler := api.InfuraRequest(payload, txInfo)
		if errHandler.Err != nil {
			err = errHandler.Err
			return
		}
		if txInfo == nil || txInfo.BlockNumber == nil {
			time.Sleep(time.Millisecond * time.Duration(i+1) * 10)
			continue
		}
		return txInfo, nil
	}
	err = errors.New("failed to download transaction info " + transactionHash + " after " + string(rune(downloadAttempts)) + " attempts")
	return
}

func getTransactionReceipt(transactionHash string) (txReceipt *api.EthTransactionReceipt, err error) {
	payload := map[string]any{
		"jsonrpc": "2.0",
		"method":  "eth_getTransactionReceipt",
		"id":      time.Now().UnixMilli(),
		"params":  []string{transactionHash},
	}
	for i := 0; i < downloadAttempts; i++ {
		txReceipt = &api.EthTransactionReceipt{}
		errHandler := api.InfuraRequest(payload, txReceipt)
		if errHandler.Err != nil {
			err = errHandler.Err
			return
		}
		if txReceipt == nil || txReceipt.BlockNumber == nil || txReceipt.GasUsed == nil {
			time.Sleep(time.Millisecond * time.Duration(i+1) * 10)
			continue
		}
		return txReceipt, nil
	}
	err = errors.New("failed to download transaction receipt " + transactionHash + " after " + string(rune(downloadAttempts)) + " attempts")
	return
}

func parseEthEventLog(eventLog *api.EthEventLog, blockchain string) *TransactionLog {
	if eventLog.Address != nil && eventLog.Data != nil {
		blockNumber, _ := helpers.HexConvertToInt(*eventLog.BlockNumber)
		logIndex, _ := helpers.HexConvertToInt(*eventLog.LogIndex)
		transactionIndex, _ := helpers.HexConvertToInt(*eventLog.TransactionIndex)
		cleanTopics := helpers.ArrayMap(eventLog.Topics, func(t string) (bool, string) {
			return true, helpers.HexRemoveLeadingZeros(t)
		}, true, "")
		txLog := &TransactionLog{}
		txLog.Blockchain = blockchain
		txLog.CreatedAt = time.Now()
		txLog.UpdatedAt = time.Now()
		txLog.TransactionHash = *eventLog.TransactionHash
		txLog.Address = *eventLog.Address
		txLog.TransactionIndex = transactionIndex
		txLog.Topics = cleanTopics
		txLog.EventId = strings.Join(cleanTopics, "-")
		txLog.BlockHash = *eventLog.BlockHash
		txLog.BlockNumber = blockNumber
		txLog.Data = *eventLog.Data
		txLog.LogIndex = logIndex
		txLog.Removed = *eventLog.Removed
		return txLog
	}
	return nil
}

func parseTransactionLogs(logs []api.EthEventLog, blockchain string) []*TransactionLog {
	txLogs := make([]*TransactionLog, 0)
	for _, log := range logs {
		txLog := parseEthEventLog(&log, blockchain)
		if txLog != nil {
			txLogs = append(txLogs, txLog)
		}
	}
	return txLogs
}

func parseTransactionInfo(transactionHash *hashes.TransactionHash, txDetails *api.EthTransaction, txReceipt *api.EthTransactionReceipt) (*TransactionInfo, []*TransactionLog) {
	var txInfo *TransactionInfo
	txLogs := make([]*TransactionLog, 0)
	if txDetails != nil && txReceipt != nil {
		txInfo = &TransactionInfo{}
		txInfo.Blockchain = transactionHash.Blockchain
		txInfo.TransactionHash = transactionHash.TransactionHash
		txInfo.BlockNumber, _ = helpers.HexConvertToInt(*txDetails.BlockNumber)
		txInfo.BlockHash = *txDetails.BlockHash
		txInfo.BlockTimestamp = transactionHash.BlockTimestamp
		if txDetails.ChainID != nil {
			txInfo.ChainID, _ = helpers.HexConvertToString(*txDetails.ChainID)
		}
		txInfo.Gas, _ = helpers.HexConvertToString(*txDetails.Gas)
		txInfo.GasUsed, _ = helpers.HexConvertToString(*txReceipt.GasUsed)
		txInfo.CumulativeGasUsed, _ = helpers.HexConvertToString(*txReceipt.CumulativeGasUsed)
		txInfo.GasPrice, _ = helpers.HexConvertToString(*txReceipt.EffectiveGasPrice)
		txInfo.From = *txDetails.From
		txInfo.To = *txDetails.To
		txInfo.Value, _ = helpers.HexConvertToString(*txDetails.Value)
		txInfo.TransactionIndex, _ = helpers.HexConvertToInt(*txDetails.TransactionIndex)
		txInfo.Input = *txDetails.Input
		txInfo.Nonce, _ = helpers.HexConvertToInt(*txDetails.Nonce)
		txInfo.R = *txDetails.R
		txInfo.S = *txDetails.S
		txInfo.V, _ = helpers.HexConvertToString(*txDetails.V)
		txInfo.Type, _ = helpers.HexConvertToString(*txDetails.Type)
		txInfo.Status, _ = helpers.HexConvertToString(*txReceipt.Status)
	}
	if txReceipt != nil {
		txLogs = parseTransactionLogs(txReceipt.Logs, transactionHash.Blockchain)
	}
	return txInfo, txLogs
}

func downloadTransactionDataByHash(txHashInput *TransactionInput) (*TransactionInfo, []*TransactionLog, error) {
	var txDetails *api.EthTransaction
	var txReceipt *api.EthTransactionReceipt
	var err error
	if txHashInput.FetchInfo {
		txDetails, err = getTransactionByHash(txHashInput.Hash.TransactionHash)
	}
	if err != nil {
		return nil, nil, err
	}
	if txHashInput.FetchLogs {
		txReceipt, err = getTransactionReceipt(txHashInput.Hash.TransactionHash)
	}
	if err != nil {
		return nil, nil, err
	}
	txInfo, tTxLogs := parseTransactionInfo(txHashInput.Hash, txDetails, txReceipt)
	return txInfo, tTxLogs, err
}

func saveLogsInDatabase(txLogs []*TransactionLog, dbInstance *mongo.Database) error {
	if txLogs != nil && len(txLogs) > 0 {
		dbCollection := database.CollectionInstance(dbInstance, &TransactionLog{})

		operations := make([]mongo.WriteModel, len(txLogs))
		for i, txLog := range txLogs {
			var filterPayload = bson.M{"address": txLog.Address, "blockchain": txLog.Blockchain, "transaction_hash": txLog.TransactionHash, "event_id": txLog.EventId}
			operations[i] = mongo.NewReplaceOneModel().SetFilter(filterPayload).SetReplacement(txLog).SetUpsert(true)
		}
		_, err := dbCollection.BulkWrite(context.Background(), operations)
		return err
	}
	return nil
}

func saveInfosInDatabase(txInfos []*TransactionInfo, dbInstance *mongo.Database) error {
	if txInfos != nil && len(txInfos) > 0 {
		dbCollection := database.CollectionInstance(dbInstance, &TransactionInfo{})

		operations := make([]mongo.WriteModel, len(txInfos))
		for i, txInfo := range txInfos {
			var filterPayload = bson.M{"transaction_hash": txInfo.TransactionHash, "blockchain": txInfo.Blockchain}
			operations[i] = mongo.NewReplaceOneModel().SetFilter(filterPayload).SetReplacement(txInfo).SetUpsert(true)
		}
		_, err := dbCollection.BulkWrite(context.Background(), operations)
		return err
	}
	return nil
}

func SaveTransactionData(txInfos []*TransactionInfo, txLogs []*TransactionLog) error {
	dbInstance, err := database.NewDatabaseConnection()
	if err != nil {
		return err
	}
	defer database.CloseDatabaseConnection(dbInstance)

	err = saveInfosInDatabase(txInfos, dbInstance)
	if err != nil {
		return err
	}
	err = saveLogsInDatabase(txLogs, dbInstance)
	return err
}

func DownloadTransactionData(hashInputs []*TransactionInput, _ *sync.WaitGroup) error {
	txInfos := make([]*TransactionInfo, 0)
	txLogs := make([]*TransactionLog, 0)
	allErrors := make([]error, 0)

	parserWg := &sync.WaitGroup{}
	dataLocker := &sync.RWMutex{}
	for _, hashInput := range hashInputs {
		parserWg.Add(1)
		go func() {
			defer parserWg.Done()
			dataLocker.Lock()
			txInfo, tTxLogs, err := downloadTransactionDataByHash(hashInput)
			if err != nil {
				allErrors = append(allErrors, err)
			} else {
				if txInfo != nil {
					txInfos = append(txInfos, txInfo)
				}
				if tTxLogs != nil && len(tTxLogs) > 0 {
					txLogs = append(txLogs, tTxLogs...)
				}
			}
			dataLocker.Unlock()
		}()
	}
	parserWg.Wait()

	if len(allErrors) > 0 {
		return allErrors[0]
	}

	_ = SaveTransactionData(txInfos, txLogs)

	//txInfos := make([]*TransactionInfo, 0)
	//txLogs := make([]*TransactionLog, 0)
	//for _, hashInput := range hashInputs {
	//	txInfo, tTxLogs, err := downloadTransactionDataByHash(hashInput)
	//	if err != nil {
	//		return err
	//	} else {
	//		if txInfo != nil {
	//			txInfos = append(txInfos, txInfo)
	//		}
	//		if tTxLogs != nil && len(tTxLogs) > 0 {
	//			txLogs = append(txLogs, tTxLogs...)
	//		}
	//	}
	//	time.Sleep(time.Millisecond * 100)
	//}
	//println(txInfos, txLogs)

	return nil
}
