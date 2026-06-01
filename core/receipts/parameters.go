package receipts

import (
	"context"
	"decentraland-data-downloader-v4/core/decentraland"
	hashesPkg "decentraland-data-downloader-v4/core/hashes"
	"decentraland-data-downloader-v4/packages/database"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getHashesFromDatabase(dbInstance *mongo.Database) (hashes map[string][]*TransactionInput, err error) {
	var dclInfo *decentraland.DclInfo
	dclInfo, err = decentraland.GetDecentralandInfo(dbInstance)
	if err != nil {
		return
	}

	hashes = make(map[string][]*TransactionInput)
	for _, blockchain := range dclInfo.Blockchains {
		filterStage := bson.D{
			{"$match", bson.D{{"blockchain", blockchain}}},
		}
		joinTxLogsStage := bson.D{
			{"$lookup", bson.D{
				{"from", "transaction_logs"}, {"localField", "transaction_hash"},
				{"foreignField", "transaction_hash"}, {"as", "log_entries"},
			}},
		}
		noLogsStage := bson.D{
			{"$match", bson.D{
				{"log_entries", bson.D{
					{"$eq", []any{}},
				}},
			}},
		}
		joinTxInfoStage := bson.D{
			{"$lookup", bson.D{
				{"from", "transaction_infos"}, {"localField", "transaction_hash"},
				{"foreignField", "transaction_hash"}, {"as", "info_entries"},
			}},
		}
		projectStage := bson.D{
			{"$project", bson.D{
				{"transaction_hash", "$$ROOT"},
				{"fetchInfo", bson.D{
					{"$cond", bson.D{
						{"if", bson.D{
							{"$gt", bson.A{
								bson.D{{"$size", "$info_entries"}},
								0,
							}},
						}},
						{"then", false},
						{"else", true},
					}},
				}},
			}},
		}
		cleanStage := bson.D{
			{"$unset", bson.A{"transaction_hash.log_entries", "transaction_hash.info_entries"}},
		}
		pipeline := mongo.Pipeline{filterStage, joinTxLogsStage, noLogsStage, joinTxInfoStage, projectStage, cleanStage}
		opts := options.Aggregate().SetAllowDiskUse(true)

		txHashCollection := database.CollectionInstance(dbInstance, &hashesPkg.TransactionHash{})
		cursor, e0 := txHashCollection.Aggregate(context.Background(), pipeline, opts)
		if e0 != nil {
			return nil, e0
		}
		rawFilteredTxHashes := make([]bson.M, 0)
		e0 = cursor.All(context.Background(), &rawFilteredTxHashes)
		if e0 != nil {
			return nil, e0
		}
		_ = cursor.Close(context.Background())
		bResults := make([]*TransactionInput, 0)
		for _, rawHash := range rawFilteredTxHashes {
			rawTxHash, _ := rawHash["transaction_hash"]
			fetchInfo, _ := rawHash["fetchInfo"]
			txHashJson, _ := json.Marshal(rawTxHash)
			txHash := &hashesPkg.TransactionHash{}
			_ = json.Unmarshal(txHashJson, txHash)
			result := &TransactionInput{
				Hash:      txHash,
				FetchLogs: true,
				FetchInfo: fetchInfo.(bool),
			}
			bResults = append(bResults, result)
		}
		hashes[blockchain] = bResults
	}
	return
}

const PartitionsNbItem = 100

func BuildParameters(dbInstance *mongo.Database) (map[string][]*TransactionInput, error) {
	allTransactionsHashes, err := getHashesFromDatabase(dbInstance)
	if err != nil {
		return nil, err
	}
	txHashesSlices := make(map[string][]*TransactionInput)
	for blockchain, transactionsHashes := range allTransactionsHashes {
		nbParts := int(math.Ceil(float64(len(transactionsHashes)) / float64(PartitionsNbItem)))
		for i := 0; i < nbParts; i++ {
			start := i * PartitionsNbItem
			end := start + PartitionsNbItem
			if end > len(transactionsHashes) {
				end = len(transactionsHashes)
			}
			txHash1 := transactionsHashes[start]
			key := fmt.Sprintf("%s_%s_%s", blockchain, txHash1.Hash.TransactionHash, txHash1.Hash.BlockTimestamp.Format(time.RFC3339))
			txHashesSlices[key] = transactionsHashes[start:end]
		}
	}
	return txHashesSlices, nil
}
