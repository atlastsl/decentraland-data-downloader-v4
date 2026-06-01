package multithread

import (
	"decentraland-data-downloader-v4/packages/logger"
	"fmt"
	"sync"
	"time"
)

type WorkerType string

const (
	WorkerTypeGetter WorkerType = "Getter"
	WorkerTypeParser WorkerType = "Parser"
)

type WorkerParserStatus string

const (
	WorkerParserStatusSuccess WorkerParserStatus = "success"
	WorkerParserStatusFailed  WorkerParserStatus = "failed"
)

type WorkerNotificationType string

const (
	WorkerNotificationTypeDone WorkerNotificationType = "done"
	WorkerNotificationTypeData WorkerNotificationType = "data"
)

type WorkerNotification struct {
	Type   WorkerNotificationType
	Task   string
	Status WorkerParserStatus
	Data   any
	Err    error
}

type WorkerNotifier func(notification *WorkerNotification, worker *Worker)

type WorkerNextCursor func(worker *Worker) (shouldWaitMoreData bool, task string, nextInput any)

type WorkerInterruptedChecker func(worker *Worker) bool

type WorkerGetterJob interface {
	FetchData(worker *Worker)
}

type WorkerParserJob interface {
	ParseData(worker *Worker, wg *sync.WaitGroup)
}

type WorkerLogger interface {
	WLogging(message string)
}

type Worker struct {
	Type       WorkerType
	Name       string
	Index      int8
	Title      string
	gJob       WorkerGetterJob
	pJob       WorkerParserJob
	Notifier   *WorkerNotifier
	NextCursor *WorkerNextCursor
	ItrChecker *WorkerInterruptedChecker
}

func NewWorker(wType WorkerType, name string, index int8, title string, gJob WorkerGetterJob, pJob WorkerParserJob, notifier *WorkerNotifier, nextCursor *WorkerNextCursor, itrChecker *WorkerInterruptedChecker) *Worker {
	return &Worker{
		Type:       wType,
		Name:       name,
		Index:      index,
		Title:      title,
		gJob:       gJob,
		pJob:       pJob,
		Notifier:   notifier,
		NextCursor: nextCursor,
		ItrChecker: itrChecker,
	}
}

func (w *Worker) loggingPrefix() string {
	return fmt.Sprintf("[%s] [%s - %d / %s (%s)]", time.Now().Format(time.RFC3339), string(w.Type), w.Index, w.Name, w.Title)
}

func (w *Worker) loggingMessage(message string) {
	logMessage := fmt.Sprintf("%s --> %s", w.loggingPrefix(), message)
	logger.Info(logMessage)
}

func (w *Worker) loggingError(preMessage string, err error) {
	logMessage := fmt.Sprintf("%s --> %s {Err = %s}", w.loggingPrefix(), preMessage, err.Error())
	logger.ErrorE(logMessage, err, false)
}

func (w *Worker) loggingStart() {
	w.loggingMessage("Started !")
}

func (w *Worker) loggingFinished() {
	w.loggingMessage("Finished !")
}

func (w *Worker) loggingDataPublished(num uint8, task string, err error) {
	if err != nil {
		logPreMessage := fmt.Sprintf("Error occurred on work (Step = #%d, Task = %s)", num, task)
		w.loggingError(logPreMessage, err)
	} else {
		logMessage := fmt.Sprintf("Data published (Step = #%d, Task = %s)", num, task)
		w.loggingMessage(logMessage)
	}
}

func (w *Worker) loggingAllDataPublished() {
	logMessage := fmt.Sprintf("Data publishing all Done !")
	w.loggingMessage(logMessage)
}

func (w *Worker) loggingTaskBeginning(shouldWaitMoreData bool, task string) {
	logMessage := ""
	if shouldWaitMoreData {
		logMessage = fmt.Sprintf("Should wait a little bit for getter to fetch data...")
	} else {
		logMessage = fmt.Sprintf("Task beginning (Task = %s)", task)
	}
	w.loggingMessage(logMessage)
}

func (w *Worker) loggingTaskFinished(task string, err error) {
	if err != nil {
		logPreMessage := fmt.Sprintf("Error occurred on work (Task = %s)", task)
		w.loggingError(logPreMessage, err)
	} else {
		logMessage := fmt.Sprintf("Task finished (Task = %s)", task)
		w.loggingMessage(logMessage)
	}
}

func (w *Worker) loggingWorkerInterrupted(reason string) {
	w.loggingMessage(fmt.Sprintf("Executable interrupted (SIG_INT, SIG_KILL) {Reason = %s}", reason))
}

func (w *Worker) LoggingExtra(message string) {
	w.loggingMessage(message)
}

func (w *Worker) LoggingError(message string, err error) {
	w.loggingError(message, err)
}

func (w *Worker) work(wg *sync.WaitGroup) {
	w.loggingStart()
	if w.Type == WorkerTypeGetter {
		w.gJob.FetchData(w)
	} else if w.Type == WorkerTypeParser {
		w.pJob.ParseData(w, wg)
	}
	w.loggingFinished()
	wg.Done()
}
