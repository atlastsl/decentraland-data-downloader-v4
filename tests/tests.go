package main

import (
	"decentraland-data-downloader-v4/core/hashes"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	hashes.Test()
}
