package hashes

import (
	"decentraland-data-downloader-v4/core/decentraland"
	"decentraland-data-downloader-v4/packages/api"
	"decentraland-data-downloader-v4/packages/database"
	"decentraland-data-downloader-v4/packages/helpers"
	"decentraland-data-downloader-v4/packages/multithread"
	"fmt"
	"reflect"
	"sync"
	"time"
)

const Argument = "hashes"

type ParametersWorker struct{}

func (x ParametersWorker) FetchData(worker *multithread.Worker) {

	var data any = true
	var err error = nil

	multithread.PublishDataNotification(worker, "-", data, err)
	multithread.PublishDoneNotification(worker)
}

type DownloaderWorker struct{}

func (x DownloaderWorker) FetchData(worker *multithread.Worker) {

	flag := false

	worker.LoggingExtra("Connecting to database...")
	databaseInstance, err := database.NewDatabaseConnection()
	if err != nil {
		worker.LoggingError("Failed to connect to database !", err)
		return
	}
	defer database.CloseDatabaseConnection(databaseInstance)
	worker.LoggingExtra("Connection to database OK!")

	worker.LoggingExtra("Get Decentraland Info data...")
	dclInfo, err := decentraland.GetDecentralandInfo(databaseInstance)
	if err != nil {
		worker.LoggingError("Failed to Get Decentraland Info data !", err)
		return
	}
	worker.LoggingExtra("Get Decentraland Info data OK!")

	topics := make([]string, 0)
	infoLogTopics := make(map[string]*decentraland.DclInfoLogTopic)
	for _, infoLogTopic := range dclInfo.LogTopics {
		if infoLogTopic.GetLogs {
			topics = append(topics, infoLogTopic.Key)
			infoLogTopics[infoLogTopic.Key] = &infoLogTopic
		}
	}

	iTopic := 0
	currentTopic := topics[iTopic]
	currentInfoLogTopic := infoLogTopics[currentTopic]
	currentBN := uint64(currentInfoLogTopic.StartBlock)

	worker.LoggingExtra("Start fetching eth events logs !")
	for !flag {

		interrupted := (*worker.ItrChecker)(worker)
		if interrupted {
			worker.LoggingExtra("Break downloader loop. Process interrupted !")
			flag = true
		} else {
			worker.LoggingExtra("Getting more data...")

			var data any = nil
			var err0 error = nil

			task := fmt.Sprintf("%s-%d", currentTopic, currentBN)

			response, err2 := DownloadEthLogs(currentInfoLogTopic, currentBN)
			if err2 != nil {
				err0 = err2
			} else {
				data = map[string]any{task: response.Events}
				if len(response.Events) > 0 {
					currentBN = response.LastBlockNumber
				} else {
					currentInfoLogTopic.StartBlock = int64(response.LastBlockNumber)
					if iTopic+1 >= len(topics) {
						flag = true
					} else {
						iTopic = iTopic + 1
						currentTopic = topics[iTopic]
						currentInfoLogTopic = infoLogTopics[currentTopic]
						currentBN = uint64(currentInfoLogTopic.StartBlock)
					}
				}
			}

			multithread.PublishDataNotification(worker, task, helpers.AnytiseData(data), err0)
			if err0 != nil {
				worker.LoggingError("Error when getting data !", err0)
				flag = true
			} else {
				worker.LoggingExtra("Sleeping 1s before getting more data...")
				time.Sleep(1 * time.Second)
			}

		}

	}

	_ = decentraland.SaveLastBlockNumber(databaseInstance, infoLogTopics)

	multithread.PublishDoneNotification(worker)

}

type SaverWorker struct{}

func (x SaverWorker) ParseData(worker *multithread.Worker, wg *sync.WaitGroup) {
	flag := false

	if worker.NextCursor != nil {
		for !flag {

			interrupted := (*worker.ItrChecker)(worker)
			if interrupted {
				flag = true
			} else {
				shouldWaitMoreData, task, nextInput := (*worker.NextCursor)(worker)
				if shouldWaitMoreData {
					time.Sleep(time.Second)
				} else if task == "" {
					flag = true
				} else if nextInput != nil {
					if reflect.TypeOf(nextInput).Kind() == reflect.Map {
						niMap := nextInput.(map[string]any)
						mainData := niMap["mainData"]
						events := mainData.([]*api.EthEventLog)

						err := SaveEthEvents(events, wg)

						multithread.PublishTaskDoneNotification(worker, task, err)

					}
				}
			}

		}
	} else {
		worker.LoggingExtra("Next Cursor is Null !!!")
	}
}

func Launch(metaverse string, nbParsers int) {

	addDataJob := &ParametersWorker{}
	mainDataJob := &DownloaderWorker{}
	parserJob := &SaverWorker{}

	workTitle := "Transaction Hashes Downloader"
	workerTitles := []string{
		"[-] Ignored Parameters Builder",
		"Eth Events Logs Downloader",
		"Eth Events Logs --> Transaction Hashes Converter",
	}
	workerDescriptions := []string{
		"[-] Ignored Parameters Builder",
		"Download all Eth events logs from Infura API",
		"Convert all Eth events logs to transaction hashes and save in database",
	}

	multithread.Launch(
		metaverse,
		addDataJob,
		mainDataJob,
		parserJob,
		nbParsers,
		workTitle,
		workerTitles,
		workerDescriptions,
	)

}
