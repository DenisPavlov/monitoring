package storage

type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

func (s *MemStorage) AddGauge(name string, value float64) error {
	s.Gauges[name] = value
	return nil
}

func (s *MemStorage) AddCounter(name string, value int64) error {
	s.Counters[name] = s.Counters[name] + value
	return nil
}

func (s *MemStorage) Gauge(name string) (float64, bool) {
	i, ok := s.Gauges[name]
	return i, ok
}

func (s *MemStorage) Counter(name string) (int64, bool) {
	i, ok := s.Counters[name]
	return i, ok
}

func (s *MemStorage) AllGauges() map[string]float64 {
	return s.Gauges
}

func (s *MemStorage) AllCounters() map[string]int64 {
	return s.Counters
}
