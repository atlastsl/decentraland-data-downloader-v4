package receipts

import "decentraland-data-downloader-v4/packages/database"

func Test() {
	databaseInstance, err := database.NewDatabaseConnection()
	if err != nil {
		panic(err)
	}
	defer database.CloseDatabaseConnection(databaseInstance)

	data, err := BuildParameters(databaseInstance)
	if err != nil {
		panic(err)
	}
	keys := make([]string, 0)
	for key := range data {
		keys = append(keys, key)
	}
	err = DownloadTransactionData(data[keys[0]], nil)
	if err != nil {
		panic(err)
	}
}
