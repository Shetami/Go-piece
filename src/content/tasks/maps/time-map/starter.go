package main

// TimeMap — хранилище «ключ → история значений по времени».
//
//   - Set(key, value, ts) записывает значение ключа на момент ts. Моменты
//     приходят в любом порядке (поздние сообщения, догрузка истории).
//     Повторный Set с тем же ts заменяет значение на этот момент.
//   - Get(key, ts) — значение, действовавшее в момент ts: запись с
//     наибольшим ts' <= ts. Если такой нет (ключа нет или все записи позже) —
//     "", false. Сложность — O(log n) по числу записей ключа.
type TimeMap struct {
	// ваши поля
}

func NewTimeMap() *TimeMap {
	return &TimeMap{}
}

func (m *TimeMap) Set(key, value string, ts int) {
	// ваш код
}

func (m *TimeMap) Get(key string, ts int) (string, bool) {
	// ваш код
	return "", false
}
