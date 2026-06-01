package hashes

import (
	"time"

	"github.com/kamva/mgm/v3"
)

type TransactionHash struct {
	mgm.DefaultModel `bson:",inline"`
	Blockchain       string    `bson:"blockchain,omitempty" json:"blockchain"`
	TransactionHash  string    `bson:"transaction_hash,omitempty" json:"transaction_hash"`
	BlockNumber      int       `bson:"block_number,omitempty" json:"block_number"`
	BlockHash        string    `bson:"block_hash,omitempty" json:"block_hash"`
	BlockTimestamp   time.Time `bson:"block_timestamp,omitempty" json:"block_timestamp"`
}
