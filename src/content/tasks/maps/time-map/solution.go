package main

import (
	"slices"
	"sort"
)

// TimeMap — хранилище «ключ → история значений по времени».
//
//   - Set(key, value, ts) записывает значение ключа на момент ts. Моменты
//     приходят в любом порядке (поздние сообщения, догрузка истории).
//     Повторный Set с тем же ts заменяет значение на этот момент.
//   - Get(key, ts) — значение, действовавшее в момент ts: запись с
//     наибольшим ts' <= ts. Если такой нет (ключа нет или все записи позже) —
//     "", false. Сложность — O(log n) по числу записей ключа.
type TimeMap struct {
	hist map[string][]version // у каждого ключа — версии, отсортированные по ts
}

type version struct {
	ts    int
	value string
}

func NewTimeMap() *TimeMap {
	return &TimeMap{hist: make(map[string][]version)}
}

func (m *TimeMap) Set(key, value string, ts int) {
	vs := m.hist[key]
	i := sort.Search(len(vs), func(i int) bool { return vs[i].ts >= ts })
	if i < len(vs) && vs[i].ts == ts {
		vs[i].value = value // тот же момент — замена, а не вторая версия
		return
	}
	// Вставка на своё место: чаще всего это конец слайса, и сдвига нет.
	m.hist[key] = slices.Insert(vs, i, version{ts, value})
}

func (m *TimeMap) Get(key string, ts int) (string, bool) {
	vs := m.hist[key]
	// Первая версия строго позже ts; нужная — перед ней.
	i := sort.Search(len(vs), func(i int) bool { return vs[i].ts > ts })
	if i == 0 {
		return "", false
	}
	return vs[i-1].value, true
}
