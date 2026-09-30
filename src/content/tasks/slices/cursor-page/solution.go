package main

import "sort"

// Item — элемент каталога; ID уникальны.
type Item struct {
	ID    string
	Title string
}

// Page отдаёт одну страницу: до limit элементов, идущих строго после курсора
// after, и курсор для следующей страницы. items отсортированы по ID по
// возрастанию. after == "" — с самого начала.
//
// Курсор — ID последнего элемента страницы. Если после страницы элементов
// больше нет, next == "". Элемента с ID == after могло уже не стать —
// страница всё равно начинается с первого ID больше after.
// Поиск начала — O(log n). append к странице не должен портить items.
// limit <= 0 — пустая страница и next == "".
func Page(items []Item, after string, limit int) (page []Item, next string) {
	if limit <= 0 {
		return nil, ""
	}
	// Первый ID строго больше курсора — работает и для удалённого курсора.
	start := sort.Search(len(items), func(i int) bool { return items[i].ID > after })
	end := min(start+limit, len(items))
	// Третий индекс: append к странице уйдёт в новый массив, а не в items.
	page = items[start:end:end]
	if end < len(items) {
		next = items[end-1].ID
	}
	return page, next
}
