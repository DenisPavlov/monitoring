package storage

type Storage interface {
	AddGauge(name string, value float64)
	AddCounter(name string, value int64)
}

type MemStorage struct {
	gauges  map[string]float64
	counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:  make(map[string]float64),
		counter: make(map[string]int64),
	}
}

func (s *MemStorage) AddGauge(name string, value float64) {
	s.gauges[name] = value
}

func (s *MemStorage) AddCounter(name string, value int64) {
	s.counter[name] = s.counter[name] + value
}

func (s *MemStorage) Gauge(name string) float64 {
	return s.gauges[name]
}

func (s *MemStorage) Counter(name string) int64 {
	return s.counter[name]
}
