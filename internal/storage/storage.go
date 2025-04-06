package storage

type Storage interface {
	AddGauge(name string, value float64) error
	AddCounter(name string, value int64) error
	Gauge(name string) (float64, bool)
	Counter(name string) (int64, bool)
	AllGauges() map[string]float64
	AllCounters() map[string]int64
}
