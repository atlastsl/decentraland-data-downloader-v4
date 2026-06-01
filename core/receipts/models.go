package receipts

import (
	"decentraland-data-downloader-v4/core/hashes"
	"time"

	"github.com/kamva/mgm/v3"
)

type TransactionLog struct {
	mgm.DefaultModel `bson:",inline"`
	Blockchain       string   `bson:"blockchain,omitempty"`
	TransactionHash  string   `bson:"transaction_hash,omitempty"`
	Address          string   `bson:"address,omitempty"`
	TransactionIndex int      `bson:"transaction_index,omitempty"`
	Topics           []string `bson:"topics,omitempty"`
	EventId          string   `bson:"event_id,omitempty"`
	BlockHash        string   `bson:"block_hash,omitempty"`
	BlockNumber      int      `bson:"block_number,omitempty"`
	Data             string   `bson:"data,omitempty"`
	LogIndex         int      `bson:"log_index,omitempty"`
	Removed          bool     `bson:"removed"`
}

type TransactionInfo struct {
	mgm.DefaultModel  `bson:",inline"`
	Blockchain        string    `bson:"blockchain,omitempty"`
	TransactionHash   string    `bson:"transaction_hash,omitempty"`
	BlockNumber       int       `bson:"block_number,omitempty"`
	BlockHash         string    `bson:"block_hash,omitempty"`
	BlockTimestamp    time.Time `bson:"block_timestamp,omitempty"`
	ChainID           string    `bson:"chain_id,omitempty"`
	Gas               string    `bson:"gas,omitempty"`
	GasUsed           string    `bson:"gas_used,omitempty"`
	CumulativeGasUsed string    `bson:"cumulative_gas_used,omitempty"`
	GasPrice          string    `bson:"gas_price,omitempty"`
	From              string    `bson:"from,omitempty"`
	To                string    `bson:"to,omitempty"`
	Value             string    `bson:"value,omitempty"`
	TransactionIndex  int       `bson:"transaction_index,omitempty"`
	Input             string    `bson:"input,omitempty"`
	Nonce             int       `bson:"nonce,omitempty"`
	R                 string    `bson:"r,omitempty"`
	S                 string    `bson:"s,omitempty"`
	V                 string    `bson:"v,omitempty"`
	Type              string    `bson:"type,omitempty"`
	Status            string    `bson:"status,omitempty"`
}

type TransactionInput struct {
	Hash      *hashes.TransactionHash
	FetchInfo bool
	FetchLogs bool
}
