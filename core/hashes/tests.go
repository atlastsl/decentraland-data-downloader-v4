package hashes

import (
	"decentraland-data-downloader-v4/core/decentraland"
	"decentraland-data-downloader-v4/packages/database"
	"decentraland-data-downloader-v4/packages/helpers"
)

func Test() {
	dbInstance, err := database.NewDatabaseConnection()
	if err != nil {
		panic(err)
	}
	defer database.CloseDatabaseConnection(dbInstance)

	dclInfo, err := decentraland.GetDecentralandInfo(dbInstance)
	if err != nil {
		panic(err)
	}

	topics := make([]string, len(dclInfo.LogTopics))
	infoLogTopics := make(map[string]*decentraland.DclInfoLogTopic, len(dclInfo.LogTopics))
	for i, infoLogTopic := range dclInfo.LogTopics {
		topics[i] = infoLogTopic.Key
		infoLogTopics[infoLogTopic.Key] = &infoLogTopic
	}

	iTopic := 4
	currentTopic := topics[iTopic]
	currentInfoLogTopic := infoLogTopics[currentTopic]
	currentBN := uint64(currentInfoLogTopic.StartBlock)
	println(currentBN)

	response, err := DownloadEthLogs(currentInfoLogTopic, currentBN)
	if err != nil {
		panic(err)
	}
	rmp := parseEthEventLog(response.Events[0])
	helpers.PrettyPrintObject(rmp)
	println(response.LastBlockNumber)
}
