package hashes

import (
	"decentraland-data-downloader-v4/core/decentraland"
	"decentraland-data-downloader-v4/packages/api"
	"decentraland-data-downloader-v4/packages/helpers"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func defaultBlockInterval(lastBlockNumber uint64) []uint64 {
	return []uint64{lastBlockNumber + 1, 0}
}

func downloadEthLogs(topic string, contracts []string, blockInterval []uint64) (events []*api.EthEventLog, errHandler *api.EthErrorHandler) {
	errHandler = &api.EthErrorHandler{}
	if len(contracts) == 0 {
		errHandler.Err = errors.New("no contracts provided")
		return
	}

	toBlock := "latest"
	if blockInterval[1] > 0 {
		toBlock = hexutil.EncodeUint64(blockInterval[1])
	}
	params := map[string]interface{}{
		"address":   contracts,
		"topics":    []string{topic},
		"fromBlock": hexutil.EncodeUint64(blockInterval[0]),
		"toBlock":   toBlock,
	}
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getLogs",
		"params":  []any{params},
		"id":      time.Now().UnixMilli(),
	}

	events = make([]*api.EthEventLog, 0)
	errHandler = api.InfuraRequest(payload, &events)
	return
}

func downloadBlockInfo(blockNumber uint64, blockchain string) (blockInfo *api.EthBlockInfo, err error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []any{hexutil.EncodeUint64(blockNumber), false},
		"id":      time.Now().UnixMilli(),
	}
	blockInfoMap := map[string]any{}
	errHandler := api.InfuraRequest(payload, &blockInfoMap)
	if errHandler.Err != nil {
		err = errHandler.Err
		return
	}
	timestampHex, ok := blockInfoMap["timestamp"]
	if !ok {
		err = errors.New("timestamp not found")
		return
	}
	timestampHexStr := timestampHex.(string)
	timestamp, _ := hexutil.DecodeUint64(timestampHexStr)
	blockInfo = &api.EthBlockInfo{
		BlockNumber:    int64(blockNumber),
		Blockchain:     blockchain,
		BlockTimestamp: time.UnixMilli(int64(timestamp * 1000)),
	}
	return
}

func downloadBlockInfoBatch(blockNumbers []uint64, blockchain string) (blockInfos []*api.EthBlockInfo, err error) {
	blockInfos = make([]*api.EthBlockInfo, 0)
	for _, blockNumber := range blockNumbers {
		var blockInfo *api.EthBlockInfo
		blockInfo, err = downloadBlockInfo(blockNumber, blockchain)
		if err != nil {
			return
		}
		blockInfos = append(blockInfos, blockInfo)
		time.Sleep(time.Millisecond * 500)
	}
	return
}

const (
	NbItemsPerBatchBlockTimestamp = 1000
)

func populateBlockTimestamp(events []*api.EthEventLog, blockchain string) (err error) {
	notProvided := make([]string, 0)
	for _, event := range events {
		if event.BlockTimestamp == nil && !slices.Contains(notProvided, *event.BlockNumber) {
			notProvided = append(notProvided, *event.BlockNumber)
		}
	}
	if len(notProvided) == 0 {
		return
	}
	btMap := make(map[string]string)
	nbBatches := int(len(notProvided)/NbItemsPerBatchBlockTimestamp) + 1
	for batchIndex := 0; batchIndex < nbBatches; batchIndex++ {
		itemStart := batchIndex * NbItemsPerBatchBlockTimestamp
		itemEnd := (batchIndex + 1) * NbItemsPerBatchBlockTimestamp
		if itemEnd > len(notProvided) {
			itemEnd = len(notProvided)
		}
		bnsStr := notProvided[itemStart:itemEnd]
		bns := make([]uint64, 0)
		for _, bnStr := range bnsStr {
			bn, _ := hexutil.DecodeUint64(bnStr)
			bns = append(bns, bn)
		}
		bqRes, e := api.FetchBlockTimestampFromBigquery(bns, blockchain)
		if e != nil {
			err = e
			return
		}
		bnsF := helpers.ArrayMap(bqRes, func(t *api.EthBlockInfo) (bool, uint64) {
			return true, uint64(t.BlockNumber)
		}, true, 0)
		bnsNf := helpers.ArrayFilter(bns, func(u uint64) bool {
			return !slices.Contains(bnsF, u)
		})
		ifRes := make([]*api.EthBlockInfo, 0)
		if len(bnsNf) > 0 {
			ifRes, e = downloadBlockInfoBatch(bnsNf, blockchain)
		}
		for _, resItem := range bqRes {
			btMap[hexutil.EncodeUint64(uint64(resItem.BlockNumber))] = hexutil.EncodeUint64(uint64(resItem.BlockTimestamp.Unix()))
		}
		for _, resItem := range ifRes {
			btMap[hexutil.EncodeUint64(uint64(resItem.BlockNumber))] = hexutil.EncodeUint64(uint64(resItem.BlockTimestamp.Unix()))
		}
	}
	for _, event := range events {
		if event.BlockTimestamp == nil {
			btm := btMap[*event.BlockNumber]
			event.BlockTimestamp = &btm
		}
	}
	return
}

func filterEthEvents(events []*api.EthEventLog, infoLogTopic *decentraland.DclInfoLogTopic) (filtered []*api.EthEventLog) {
	if infoLogTopic.FilterLogs == "" {
		return events
	}
	filterPars := strings.Split(infoLogTopic.FilterLogs, ".")
	filterBy, filterByIdxStr := filterPars[0], filterPars[1]
	filterByIdx, _ := strconv.Atoi(filterByIdxStr)
	filtered = make([]*api.EthEventLog, 0)
	for _, event := range events {
		filterValue := ""
		if filterBy == "data" {
			evtData := *event.Data
			evtData = evtData[2:]
			filterValue = evtData[(filterByIdx-1)*64 : (filterByIdx-1)*64+64]
		} else if filterBy == "topics" {
			filterValue = event.Topics[filterByIdx-1]
		}
		if filterValue == "" {
			filtered = append(filtered, event)
		} else {
			if !strings.HasPrefix(filterValue, "0x") {
				filterValue = "0x" + filterValue
			}
			filterValue = helpers.HexRemoveLeadingZeros(filterValue)
			if slices.Contains(infoLogTopic.FilterValue, filterValue) {
				filtered = append(filtered, event)
			}
		}
	}
	return
}

type DownloadEthLogsResponse struct {
	Events          []*api.EthEventLog
	LastBlockNumber uint64
}

func DownloadEthLogs(infoLogTopic *decentraland.DclInfoLogTopic, lastBlockNumber uint64) (response *DownloadEthLogsResponse, err error) {
	defBlockInterval := defaultBlockInterval(lastBlockNumber)
	events, errHandler := downloadEthLogs(infoLogTopic.Topic, infoLogTopic.Contracts, defBlockInterval)
	if errHandler.Err != nil {
		return nil, errHandler.Err
	} else if errHandler.BlockInterval != nil && len(errHandler.BlockInterval) > 0 {
		events, errHandler = downloadEthLogs(infoLogTopic.Topic, infoLogTopic.Contracts, errHandler.BlockInterval)
	}
	if errHandler.Err != nil {
		return nil, errHandler.Err
	}
	nLastBlockNumber := lastBlockNumber
	if len(events) > 0 {
		tmp := *events[len(events)-1].BlockNumber
		nLastBlockNumber, err = hexutil.DecodeUint64(tmp)
	}
	events = filterEthEvents(events, infoLogTopic)
	for _, event := range events {
		event.Blockchain = &infoLogTopic.Blockchain
	}
	err = populateBlockTimestamp(events, infoLogTopic.Blockchain)
	if err != nil {
		return nil, err
	}
	response = &DownloadEthLogsResponse{
		Events:          events,
		LastBlockNumber: nLastBlockNumber,
	}
	return
}
