package datapipe

type ScannerStat struct {
	ScanCount    int
	MessageCount int
	StatusCount  map[string]int
}
