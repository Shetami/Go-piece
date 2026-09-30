package main

import (
	"slices"
	"sort"
)

// Entry — строка таблицы лидеров.
type Entry struct {
	Name  string
	Score int
}

// Board хранит лучшие n результатов. Поля — на ваше усмотрение.
type Board struct {
	n       int
	entries []Entry        // по убыванию Score, при равенстве — по времени
	score   map[string]int // счёт тех, кто сейчас в таблице
}

// NewBoard создаёт таблицу на n мест. n <= 0 — паника.
func NewBoard(n int) *Board {
	if n <= 0 {
		panic("NewBoard: n должно быть больше нуля")
	}
	return &Board{n: n, entries: make([]Entry, 0, n+1), score: make(map[string]int)}
}

// Submit сообщает новый результат игрока.
//   - Игрок в таблице бывает не больше одного раза. Если он уже в таблице
//     со счётом не ниже нового — ничего не меняется; если ниже — старая
//     запись заменяется новой.
//   - Таблица упорядочена по убыванию Score; при равенстве выше тот, кто
//     набрал этот счёт раньше.
//   - Если мест больше нет, последняя запись вылетает. Результат, равный
//     последнему в полной таблице, в неё не попадает. Вылетевший игрок
//     забывается и может вернуться с любым счётом.
//
// Место ищется двоичным поиском.
func (b *Board) Submit(name string, score int) {
	if old, ok := b.score[name]; ok {
		if score <= old {
			return
		}
		// Старая запись среди равных по счёту: двоичный поиск до начала
		// блока с этим счётом, дальше — по имени.
		i := sort.Search(len(b.entries), func(k int) bool { return b.entries[k].Score <= old })
		for b.entries[i].Name != name {
			i++
		}
		b.entries = slices.Delete(b.entries, i, i+1)
		delete(b.score, name)
	}
	// Первое место, где счёт строго меньше: равные остаются выше новичка.
	i := sort.Search(len(b.entries), func(k int) bool { return b.entries[k].Score < score })
	if i >= b.n {
		return
	}
	b.entries = slices.Insert(b.entries, i, Entry{name, score})
	b.score[name] = score
	if len(b.entries) > b.n {
		last := b.entries[b.n]
		delete(b.score, last.Name) // иначе вылетевший не сможет вернуться
		b.entries = slices.Delete(b.entries, b.n, b.n+1)
	}
}

// Top возвращает копию таблицы, от первого места к последнему.
func (b *Board) Top() []Entry {
	return slices.Clone(b.entries)
}
