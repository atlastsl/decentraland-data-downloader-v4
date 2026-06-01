package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var consoleOutLogger *log.Logger
var consoleErrLogger *log.Logger
var fileOutLogger *log.Logger
var fileErrLogger *log.Logger

var fileOut *os.File
var fileErr *os.File

func CloseLogger() {
	if fileOut != nil {
		err := fileOut.Close()
		if err != nil {
			return
		}
	}
	if fileErr != nil {
		err := fileErr.Close()
		if err != nil {
			return
		}
	}
}

func createLoggers(metaverse string, purpose string) {
	mainLogsDir := os.Getenv("LOGS_DIR")
	_purpose := strings.ReplaceAll(strings.ToLower(purpose), " ", "_")
	currentLogsDir := fmt.Sprintf("%s%s%s_%s_%d", mainLogsDir, string(os.PathSeparator), metaverse, _purpose, time.Now().UnixMilli())

	fileOutPath, fileErrPath := "", ""
	e0 := os.Mkdir(currentLogsDir, 0777)
	if e0 == nil {
		fileOutPath = currentLogsDir + string(os.PathSeparator) + "out.log"
		fileErrPath = currentLogsDir + string(os.PathSeparator) + "err.log"
	}

	if fileOutPath != "" {
		f, err := os.OpenFile(fileOutPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Error Opening Out Log File: %v", err)
		} else {
			fileOut = f
		}
	}
	if fileErrPath != "" {
		f, err := os.OpenFile(fileErrPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Error Opening Err Log File: %v", err)
		} else {
			fileErr = f
		}
	}

	prefix := "Metaverses Estate Data Downloader"
	consoleOutLogger = log.New(os.Stdout, prefix+" [Info] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	consoleErrLogger = log.New(os.Stdout, prefix+" [Error] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	if fileOut != nil {
		fileOutLogger = log.New(fileOut, prefix+" [Info] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile|log.LUTC)
	}
	if fileErr != nil {
		fileErrLogger = log.New(fileErr, prefix+" [Error] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile|log.LUTC)
	}
}

func InitializeLogger(metaverse string, purpose string) {
	createLoggers(metaverse, purpose)
}

func Info(message string) {
	if consoleOutLogger != nil {
		consoleOutLogger.Println(message)
	} else {
		log.Println(message)
	}
	if fileOutLogger != nil {
		fileOutLogger.Println(message)
	}
}

func Infof(format string, v ...any) {
	consoleOutLogger.Printf(format, v)
	if fileOutLogger != nil {
		fileOutLogger.Printf(format, v)
	}
}

func Error(message string) {
	if consoleErrLogger != nil {
		consoleErrLogger.Println(message)
	} else {
		log.Println(message)
	}
	if fileErrLogger != nil {
		fileErrLogger.Println(message)
	}
}

func Errore(err error, _panic bool) {
	if consoleErrLogger != nil {
		consoleErrLogger.Println(err)
	} else {
		log.Println(err)
	}
	if fileErrLogger != nil {
		fileErrLogger.Println(err)
	}
	if true {
		panic(err)
	}
}

func ErrorE(message string, err error, _panic bool) {
	Error(message)
	if err != nil {
		Errore(err, _panic)
	}
}
