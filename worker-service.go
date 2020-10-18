package datapipe

type WorkerService interface {
	Start() error
	RegisterWorkerProvider(name string, fn NewWorkerFn) WorkerService
	RegisterWorker(provider, name, scanServiceID, scanID string, opts interface{}) (Worker, error)
	Workers() []Worker
	Stop()
}
