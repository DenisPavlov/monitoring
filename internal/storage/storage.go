package storage

type Storage interface {
	addGauge(name string, value float64)
	addCounter(name string, value int64)
}

func AddGauge(s Storage, name string, value float64) {
	s.addGauge(name, value)
}

func AddCounter(s Storage, name string, value int64) {
	s.addCounter(name, value)
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

func (s *MemStorage) addGauge(name string, value float64) {
	s.gauges[name] = value
}

func (s *MemStorage) addCounter(name string, value int64) {
	s.counter[name] = s.counter[name] + value
}
