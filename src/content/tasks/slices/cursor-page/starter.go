package main

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
	// ваш код
	return nil, ""
}
