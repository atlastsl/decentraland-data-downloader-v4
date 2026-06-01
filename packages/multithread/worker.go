package multithread

import (
	"decentraland-data-downloader-v4/packages/helpers"
	"decentraland-data-downloader-v4/packages/logger"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"runtime/debug"
	"slices"
	"sync"
	"syscall"
)

var interruptHandler chan os.Signal

var (
	interruptedLocker sync.RWMutex
	interrupted       bool
	interruptReason   string
)

var (
	addDataLocker sync.RWMutex
	addData       any
	addDataDone   bool
	addDataIndex  uint8 = 0
)

var (
	mainDataLocker sync.RWMutex
	mainData       map[string]any
	mainDataDone   bool
	mainDataIndex  uint8 = 0
)

var (
	tLocker sync.RWMutex
	tasks   []string
)

var (
	tDoneLocker    sync.RWMutex
	tasksSucceeded []string
	tasksFailed    []string
	errorMessages  []string
)

var AdditionalDataNotifier WorkerNotifier = func(notification *WorkerNotification, worker *Worker) {
	addDataLocker.Lock()
	defer addDataLocker.Unlock()
	if notification != nil {
		if notification.Type == WorkerNotificationTypeData {
			addDataIndex++
			if notification.Status == WorkerParserStatusSuccess {
				addData = notification.Data
			}
			worker.loggingDataPublished(addDataIndex, notification.Task, notification.Err)
		} else if notification.Type == WorkerNotificationTypeDone {
			addDataDone = true
			worker.loggingAllDataPublished()
		}
	}
}

var MainDataNotifier WorkerNotifier = func(notification *WorkerNotification, worker *Worker) {
	mainDataLocker.Lock()
	tLocker.Lock()
	tDoneLocker.RLock()
	defer tDoneLocker.RUnlock()
	defer tLocker.Unlock()
	defer mainDataLocker.Unlock()
	if mainData == nil {
		mainData = make(map[string]any)
	}
	if tasks == nil {
		tasks = make([]string, 0)
	}
	if notification != nil {
		if notification.Type == WorkerNotificationTypeData {
			mainDataIndex++
			if notification.Status == WorkerParserStatusSuccess && notification.Data != nil {
				if reflect.TypeOf(notification.Data).Kind() == reflect.Map {
					notifData := notification.Data.(map[string]interface{})
					notifDataTasksRaw, ok := notifData["tasks"]
					if ok {
						notifDataTasks := notifDataTasksRaw.([]string)
						notifDataMap := helpers.AnytiseData(notifData["data"]).(map[string]interface{})
						for _, key := range notifDataTasks {
							if (tasksSucceeded == nil || !slices.Contains(tasksSucceeded, key)) && (tasksFailed == nil || !slices.Contains(tasksFailed, key)) {
								tasks = append(tasks, key)
								mainData[key] = notifDataMap[key]
							}
						}
					} else {
						for key, _data := range notification.Data.(map[string]any) {
							if _, ok := mainData[key]; !ok {
								if (tasksSucceeded == nil || !slices.Contains(tasksSucceeded, key)) && (tasksFailed == nil || !slices.Contains(tasksFailed, key)) {
									tasks = append(tasks, key)
									mainData[key] = _data
								}
							}
						}
					}
				} else if reflect.TypeOf(notification.Data).Kind() == reflect.Slice {
					for _, _key := range notification.Data.([]any) {
						if reflect.TypeOf(_key).Kind() == reflect.String {
							key := _key.(string)
							if (tasksSucceeded == nil || !slices.Contains(tasksSucceeded, key)) && (tasksFailed == nil || !slices.Contains(tasksFailed, key)) {
								tasks = append(tasks, key)
								mainData[key] = key
							}
						}
					}
				}
			}
			worker.loggingDataPublished(mainDataIndex, notification.Task, notification.Err)
		} else if notification.Type == WorkerNotificationTypeDone {
			mainDataDone = true
			worker.loggingAllDataPublished()
		}
	}
}

var ParserWorkerNotifier WorkerNotifier = func(notification *WorkerNotification, worker *Worker) {
	tDoneLocker.Lock()
	defer tDoneLocker.Unlock()
	if notification != nil {
		if notification.Type == WorkerNotificationTypeData {
			if notification.Status == WorkerParserStatusSuccess {
				if tasksSucceeded == nil {
					tasksSucceeded = make([]string, 0)
				}
				tasksSucceeded = append(tasksSucceeded, notification.Task)
			} else if notification.Status == WorkerParserStatusFailed {
				if tasksFailed == nil {
					tasksFailed = make([]string, 0)
				}
				tasksFailed = append(tasksFailed, notification.Task)
				if notification.Err != nil {
					if errorMessages == nil {
						errorMessages = make([]string, 0)
					}
					errorMessages = append(errorMessages, notification.Err.Error())
				}
			}
			worker.loggingTaskFinished(notification.Task, notification.Err)
			delete(mainData, notification.Task)
		}
	}
}

var ParserWorkerNextCursor WorkerNextCursor = func(worker *Worker) (shouldWaitMoreData bool, task string, nextInput any) {
	tLocker.Lock()
	mainDataLocker.RLock()
	defer mainDataLocker.RUnlock()
	defer tLocker.Unlock()
	shouldWaitMoreData = true
	task = ""
	nextInput = nil
	if addDataDone {
		if tasks != nil && len(tasks) > 0 {
			shouldWaitMoreData = false
			outTask := tasks[0]
			tasks = tasks[1:]
			nextData := make(map[string]any)
			nextData["addData"] = addData
			nextData["mainData"] = mainData[outTask]
			task = outTask
			nextInput = nextData
		} else {
			shouldWaitMoreData = !mainDataDone
		}
	}
	worker.loggingTaskBeginning(shouldWaitMoreData, task)
	return
}

var WorkerInterruptChecker WorkerInterruptedChecker = func(worker *Worker) bool {
	interruptedLocker.RLock()
	defer interruptedLocker.RUnlock()
	if interrupted {
		worker.loggingWorkerInterrupted(interruptReason)
	}
	return interrupted
}

func launch(workers []*Worker) {
	logger.Info(fmt.Sprintf("[Workers (%d workers)] - Launch workers", len(workers)))
	var workerWg sync.WaitGroup
	for i := 0; i < len(workers); i++ {
		workerWg.Add(1)
		go workers[i].work(&workerWg)
	}
	go listenInterrupt()
	workerWg.Wait()
	logger.Info(fmt.Sprintf("[Workers (%d workers)] - Workers jobs all Done", len(workers)))
}

func Launch(metaverse string, addDataJob, mainDataJob WorkerGetterJob, writerJob WorkerParserJob, nbParsers int, workTitle string, workerTitles []string, workerDescriptions []string) {
	logger.InitializeLogger(metaverse, workTitle)

	logger.Info(fmt.Sprintf("%s", workTitle))
	logger.Info("--------------------------------------------------------------")
	logger.Info(fmt.Sprintf("%s - Build workers", workTitle))

	workers := make([]*Worker, 0)
	gAddDataWorker := NewWorker(
		WorkerTypeGetter,
		fmt.Sprintf("<%s> %s", metaverse, workerTitles[0]),
		1,
		fmt.Sprintf("<%s> %s", metaverse, workerDescriptions[0]),
		addDataJob,
		nil,
		&AdditionalDataNotifier,
		nil,
		&WorkerInterruptChecker,
	)
	gMainDataWorker := NewWorker(
		WorkerTypeGetter,
		fmt.Sprintf("<%s> %s", metaverse, workerTitles[1]),
		2,
		fmt.Sprintf("<%s> %s", metaverse, workerDescriptions[1]),
		mainDataJob,
		nil,
		&MainDataNotifier,
		nil,
		&WorkerInterruptChecker,
	)
	workers = append(workers, gAddDataWorker)
	workers = append(workers, gMainDataWorker)
	for i := 0; i < nbParsers; i++ {
		parserWorkerI := NewWorker(
			WorkerTypeParser,
			fmt.Sprintf("<%s> %s", metaverse, workerTitles[2]),
			int8(i+1),
			fmt.Sprintf("<%s> %s", metaverse, workerDescriptions[2]),
			nil,
			writerJob,
			&ParserWorkerNotifier,
			&ParserWorkerNextCursor,
			&WorkerInterruptChecker,
		)
		workers = append(workers, parserWorkerI)
	}

	if len(workers) > 0 {
		logger.Info(fmt.Sprintf("%s - Launch workers", workTitle))
		launch(workers)
	} else {
		logger.Info(fmt.Sprintf("%s - No workers built", workTitle))
	}

	logger.Info("--------------------------------------------------------------")
	logger.Info(fmt.Sprintf("%s - DONE", workTitle))
	logger.Info("--------------------------------------------------------------")

	forceInterrupt()
}

func closeApp(closeMsg *string) {
	interruptedLocker.Lock()
	defer interruptedLocker.Unlock()
	interruptReason = *closeMsg
	interrupted = true
	logger.CloseLogger()
	os.Exit(0)
}

func interruptHandlerFunc() {
	sig := <-interruptHandler
	sigStr := sig.String()
	closeApp(&sigStr)
}

func listenInterrupt() {
	interruptHandler = make(chan os.Signal)
	signal.Notify(interruptHandler, os.Interrupt, os.Kill, syscall.SIGTERM)
	interruptHandlerFunc()
}

func forceInterrupt() {
	interruptHandler <- os.Interrupt
}

func Recovery() {
	if r := recover(); r != nil {
		message := fmt.Sprintf("App crashed [Error = %s]", r)
		logger.Info(message)
		fmt.Println(string(debug.Stack()))
		closeApp(&message)
	}
}
