package main

import (
	"slices"
	"strings"
	"unicode"
)

// Index — инвертированный индекс по документам.
//
// Слова: текст режется по всему, что не буква и не цифра; регистр не важен
// ("Go," и "GO" — одно слово "go"); кириллица — тоже буквы.
//
//   - Add индексирует документ id. Повторный Add того же id заменяет
//     старый текст целиком: по словам старого текста документ больше не ищется.
//   - Remove убирает документ из индекса (отсутствующий — не ошибка).
//   - Search возвращает id документов, в которых есть ВСЕ слова запроса,
//     по возрастанию. Пустой запрос (нет ни одного слова) — пустой результат.
//     Результат — новый слайс, который вызывающий может менять.
//   - Words — сколько разных слов сейчас в индексе. Слово, которого не осталось
//     ни в одном документе, не считается и не хранится.
type Index struct {
	postings map[string]map[int]struct{} // слово → документы
	docs     map[int][]string            // документ → его уникальные слова
}

func NewIndex() *Index {
	return &Index{postings: make(map[string]map[int]struct{}), docs: make(map[int][]string)}
}

// terms возвращает уникальные слова текста.
func terms(text string) []string {
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	slices.Sort(words)
	return slices.Compact(words)
}

func (ix *Index) Add(id int, text string) {
	ix.Remove(id) // повторный Add: сначала убираем старые слова документа
	ws := terms(text)
	for _, w := range ws {
		set := ix.postings[w]
		if set == nil {
			set = make(map[int]struct{})
			ix.postings[w] = set
		}
		set[id] = struct{}{}
	}
	ix.docs[id] = ws
}

func (ix *Index) Remove(id int) {
	ws, ok := ix.docs[id]
	if !ok {
		return
	}
	delete(ix.docs, id)
	for _, w := range ws {
		set := ix.postings[w]
		delete(set, id)
		if len(set) == 0 {
			delete(ix.postings, w) // пустые списки не копим
		}
	}
}

func (ix *Index) Search(query string) []int {
	qs := terms(query)
	if len(qs) == 0 {
		return nil
	}
	sets := make([]map[int]struct{}, 0, len(qs))
	for _, w := range qs {
		set, ok := ix.postings[w]
		if !ok {
			return nil // хоть одного слова нет нигде — пересечение пусто
		}
		sets = append(sets, set)
	}
	// Пересекаем, начиная с самого короткого списка: перебор идёт по нему.
	slices.SortFunc(sets, func(a, b map[int]struct{}) int { return len(a) - len(b) })
	var res []int
	for id := range sets[0] {
		inAll := true
		for _, s := range sets[1:] {
			if _, ok := s[id]; !ok {
				inAll = false
				break
			}
		}
		if inAll {
			res = append(res, id)
		}
	}
	slices.Sort(res)
	return res
}

func (ix *Index) Words() int { return len(ix.postings) }
