package hashes

import (
	"context"
	"decentraland-data-downloader-v4/packages/api"
	"decentraland-data-downloader-v4/packages/database"
	"decentraland-data-downloader-v4/packages/helpers"
	"slices"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func parseEthEventLog(eventLog *api.EthEventLog) *TransactionHash {
	blockNumber, _ := helpers.HexConvertToInt(*eventLog.BlockNumber)
	blockTimestamp, _ := helpers.HexConvertToInt(*eventLog.BlockTimestamp)
	txHash := &TransactionHash{}
	txHash.Blockchain = *eventLog.Blockchain
	txHash.TransactionHash = *eventLog.TransactionHash
	txHash.BlockTimestamp = time.Unix(int64(blockTimestamp), 0)
	txHash.BlockHash = *eventLog.BlockHash
	txHash.BlockNumber = blockNumber
	txHash.CreatedAt = txHash.BlockTimestamp
	txHash.UpdatedAt = time.Now()
	return txHash
}

func filterEthEventLogs(eventLogs []*api.EthEventLog) []*api.EthEventLog {
	filtered := make([]*api.EthEventLog, 0)
	hashes := make([]string, 0)
	for _, log := range eventLogs {
		if !slices.Contains(hashes, *log.TransactionHash) {
			filtered = append(filtered, log)
		}
	}
	return filtered
}

func saveHashesInDatabase(transactionHashes []*TransactionHash) error {
	dbInstance, err := database.NewDatabaseConnection()
	if err != nil {
		return err
	}
	defer database.CloseDatabaseConnection(dbInstance)

	if transactionHashes != nil && len(transactionHashes) > 0 {
		dbCollection := database.CollectionInstance(dbInstance, &TransactionHash{})

		operations := make([]mongo.WriteModel, len(transactionHashes))
		for i, txHash := range transactionHashes {
			var filterPayload = bson.M{"blockchain": txHash.Blockchain, "transaction_hash": txHash.TransactionHash}
			var updatePayload = bson.M{
				"$set": bson.M{
					"blockchain":       txHash.Blockchain,
					"transaction_hash": txHash.TransactionHash,
					"block_timestamp":  txHash.BlockTimestamp,
					"block_number":     txHash.BlockNumber,
					"block_hash":       txHash.BlockHash,
					"created_at":       txHash.CreatedAt,
					"updated_at":       txHash.UpdatedAt,
				},
			}
			operations[i] = mongo.NewUpdateOneModel().SetFilter(filterPayload).SetUpdate(updatePayload).SetUpsert(true)
		}
		_, err = dbCollection.BulkWrite(context.Background(), operations)
	}
	return err
}

func SaveEthEvents(eventsLogs []*api.EthEventLog, wg *sync.WaitGroup) error {
	filtered := filterEthEventLogs(eventsLogs)

	transactionHashes := helpers.ArrayMap(filtered, func(t *api.EthEventLog) (bool, *TransactionHash) {
		return true, parseEthEventLog(t)
	}, true, nil)

	wg.Add(1)
	go func() {
		_ = saveHashesInDatabase(transactionHashes)
		wg.Done()
	}()

	return nil
}
