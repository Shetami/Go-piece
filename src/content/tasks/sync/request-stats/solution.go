package main

import "sync/atomic"

// Stats — счётчики HTTP-запросов, в которые пишут все обработчики сервера
// одновременно. Нулевое значение готово к работе. Мьютекс не нужен.
type Stats struct {
	requests     atomic.Int64
	clientErrors atomic.Int64
	serverErrors atomic.Int64
	bytes        atomic.Int64
}

// Snapshot — значения счётчиков на момент вызова.
type Snapshot struct {
	Requests     int64 // всего запросов
	ClientErrors int64 // ответы 4xx
	ServerErrors int64 // ответы 5xx
	Bytes        int64 // сумма размеров ответов
}

// Record учитывает один запрос: код ответа и размер тела в байтах.
func (s *Stats) Record(status int, bytes int) {
	s.requests.Add(1)
	s.bytes.Add(int64(bytes))
	switch {
	case status >= 500 && status <= 599:
		s.serverErrors.Add(1)
	case status >= 400 && status <= 499:
		s.clientErrors.Add(1)
	}
}

func (s *Stats) Snapshot() Snapshot {
	// Каждое поле читается атомарно, но снимок в целом — нет:
	// между чтениями могут прийти новые запросы. Для метрик это нормально.
	return Snapshot{
		Requests:     s.requests.Load(),
		ClientErrors: s.clientErrors.Load(),
		ServerErrors: s.serverErrors.Load(),
		Bytes:        s.bytes.Load(),
	}
}

// ErrorRate — доля ответов 5xx среди всех запросов, от 0 до 1.
func (s *Stats) ErrorRate() float64 {
	// Сначала ошибки, потом запросы: запросы учитываются в Record раньше
	// ошибок, поэтому доля не вылезет за 1.
	errs := s.serverErrors.Load()
	total := s.requests.Load()
	if total == 0 {
		return 0 // не 0/0 = NaN
	}
	return float64(errs) / float64(total)
}
