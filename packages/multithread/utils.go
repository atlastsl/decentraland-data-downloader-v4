package multithread

func PublishDataNotification(worker *Worker, task string, data any, err error) {
	dataNotification := WorkerNotification{
		Type: WorkerNotificationTypeData,
		Task: task,
	}
	if err != nil {
		dataNotification.Status = WorkerParserStatusFailed
		dataNotification.Err = err
	} else {
		dataNotification.Status = WorkerParserStatusSuccess
		dataNotification.Data = data
		dataNotification.Err = nil
	}
	if worker.Notifier != nil {
		(*worker.Notifier)(&dataNotification, worker)
	}
}

func PublishDoneNotification(worker *Worker) {
	doneNotification := WorkerNotification{
		Type: WorkerNotificationTypeDone,
	}
	if worker.Notifier != nil {
		(*worker.Notifier)(&doneNotification, worker)
	}
}

func PublishTaskDoneNotification(worker *Worker, task string, err error) {
	dataNotification := WorkerNotification{
		Type: WorkerNotificationTypeData,
		Task: task,
	}
	if err != nil {
		dataNotification.Status = WorkerParserStatusFailed
		dataNotification.Err = err
	} else {
		dataNotification.Status = WorkerParserStatusSuccess
		dataNotification.Err = nil
	}
	if worker.Notifier != nil {
		(*worker.Notifier)(&dataNotification, worker)
	}
}
