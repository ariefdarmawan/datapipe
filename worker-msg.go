package datapipe

type WorkerMessage struct {
	ID           string
	ServiceID    string
	ScannerID    string
	ScanProvider string
	Data         []byte
}
