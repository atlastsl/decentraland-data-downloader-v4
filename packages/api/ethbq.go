package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type EthBlockInfo struct {
	Blockchain     string
	BlockNumber    int64     `bigquery:"block_number"`
	BlockTimestamp time.Time `bigquery:"block_timestamp"`
}

func FetchBlockTimestampFromBigquery(blockNumbers []uint64, blockchain string) (timestamps []*EthBlockInfo, err error) {
	projectId := os.Getenv("ETHEREUM_ETL_PROJECT_ID")
	credsFile := os.Getenv("BIG_QUERY_CREDENTIALS_FILE")
	client, err := bigquery.NewClient(context.Background(), projectId, option.WithCredentialsFile(credsFile))
	if err != nil {
		return nil, err
	}
	queryStr := fmt.Sprintf(
		`
		SELECT number as block_number, timestamp as block_timestamp
		FROM bigquery-public-data.crypto_%s.blocks
		WHERE number IN UNNEST(@block_numbers)
		ORDER BY block_number ASC
	`, blockchain)
	query := client.Query(queryStr)
	blockNumbersInt64 := make([]int64, len(blockNumbers))
	for i, v := range blockNumbers {
		blockNumbersInt64[i] = int64(v)
	}

	query.Parameters = []bigquery.QueryParameter{
		{Name: "block_numbers", Value: blockNumbersInt64},
	}
	it, err := query.Read(context.Background())
	if err != nil {
		return nil, err
	}
	blockInfos := make([]*EthBlockInfo, 0)
	for {
		blockInfo := &EthBlockInfo{}
		err = it.Next(blockInfo)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		blockInfos = append(blockInfos, blockInfo)
	}
	return blockInfos, nil
}
