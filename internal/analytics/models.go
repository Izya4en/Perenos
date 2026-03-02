package analytics

type PerformanceMetrics struct {
	TotalTransactions      int
	TotalThroughputAmount  int
	AverageLoadingPercent  float64
	LastServiceCriticality bool
}
