package main

// Stats — счётчики HTTP-запросов, в которые пишут все обработчики сервера
// одновременно. Нулевое значение готово к работе. Мьютекс не нужен.
type Stats struct {
	// ваши поля
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
	// ваш код
}

func (s *Stats) Snapshot() Snapshot {
	// ваш код
	return Snapshot{}
}

// ErrorRate — доля ответов 5xx среди всех запросов, от 0 до 1.
func (s *Stats) ErrorRate() float64 {
	// ваш код
	return 0
}
