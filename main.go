package main

import (
	"decentraland-data-downloader-v4/core/decentraland"
	"decentraland-data-downloader-v4/core/hashes"
	"decentraland-data-downloader-v4/core/receipts"
	"decentraland-data-downloader-v4/packages/multithread"
	"flag"
	"log"
	"os"
	"slices"
	"time"

	"github.com/joho/godotenv"
)

func usage() {
	log.Println("Usage: dcldtdl4 [-m <module_name> (hashes | receipts | operations)] [-n <nb_savers>] [-c envFilePath]")
	flag.PrintDefaults()
}

func showUsageAndExit(exitCode int) {
	usage()
	os.Exit(exitCode)
}

func readFlags() (*string, *int, bool) {
	var dataType = flag.String("m", "", "Module name (hashes | receipts | operations)")
	var nbSavers = flag.Int("n", 1, "Nb Savers/Downloader (>0)")
	var envFilePath = flag.String("c", ".env", "Env File Path")
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	// go run main.go -m hashes -n 1
	// go run main.go -m receipts -n 1
	// go run main.go -m operations -n 1

	if *dataType == "" {
		showUsageAndExit(0)
		return nil, nil, false
	}
	if !slices.Contains([]string{hashes.Argument, receipts.Argument}, *dataType) {
		showUsageAndExit(0)
		return nil, nil, false
	}
	if *nbSavers < 0 {
		showUsageAndExit(0)
		return nil, nil, false
	}
	err := godotenv.Load(*envFilePath)
	if err != nil {
		log.Fatalf("Fail to load %s env file", *envFilePath)
		return nil, nil, false
	}

	return dataType, nbSavers, true
}

func main() {
	defer multithread.Recovery()
	dataType, nbParsers, ok := readFlags()
	if !ok {
		os.Exit(0)
	}
	if *dataType == hashes.Argument {
		hashes.Launch(decentraland.Decentraland, *nbParsers)
	} else if *dataType == receipts.Argument {
		receipts.Launch(decentraland.Decentraland, *nbParsers)
	}
	if os.Getenv("ENVIRONMENT") == "server" {
		time.Sleep(time.Hour * 24)
	}
}
