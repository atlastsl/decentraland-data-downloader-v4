package main

import (
	"decentraland-data-downloader-v4/core/receipts"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	receipts.Test()
}
