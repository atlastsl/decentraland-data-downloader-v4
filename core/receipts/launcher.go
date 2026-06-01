package receipts

import (
	"decentraland-data-downloader-v4/packages/database"
	"decentraland-data-downloader-v4/packages/helpers"
	"decentraland-data-downloader-v4/packages/multithread"
	"reflect"
	"sync"
	"time"
)

const Argument = "receipts"

type UselessWorker struct{}

func (x UselessWorker) FetchData(worker *multithread.Worker) {

	var data any = true
	var err error = nil

	multithread.PublishDataNotification(worker, "-", data, err)
	multithread.PublishDoneNotification(worker)
}

type ParametersWorker struct{}

func (x ParametersWorker) FetchData(worker *multithread.Worker) {

	worker.LoggingExtra("Connecting to database...")
	databaseInstance, err := database.NewDatabaseConnection()
	if err != nil {
		worker.LoggingError("Failed to connect to database !", err)
		return
	}
	defer database.CloseDatabaseConnection(databaseInstance)
	worker.LoggingExtra("Connection to database OK!")

	worker.LoggingExtra("Fetching transaction hashes from database...")
	data, err := BuildParameters(databaseInstance)
	worker.LoggingExtra("Fetching transaction hashes from database OK. Publishing data...")

	multithread.PublishDataNotification(worker, "-", helpers.AnytiseData(data), err)
	multithread.PublishDoneNotification(worker)

}

type DownloaderWorker struct{}

func (x DownloaderWorker) ParseData(worker *multithread.Worker, wg *sync.WaitGroup) {
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

						err := DownloadTransactionData(mainData.([]*TransactionInput), wg)

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

	addDataJob := &UselessWorker{}
	mainDataJob := &ParametersWorker{}
	parserJob := &DownloaderWorker{}

	workTitle := "Transactions Info/Receipts Downloader"
	workerTitles := []string{
		"[-] Useless worker",
		"Transactions Hashes Builder",
		"Transaction Hash --> Transaction info downloader & Saver",
	}
	workerDescriptions := []string{
		"[-] Useless worker",
		"Get all transactions hashes from database",
		"Fetch transaction infos from Infura by transaction hash",
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
