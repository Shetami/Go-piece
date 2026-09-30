package main

// Store — хранилище ключ-значение с историей изменений.
// Не безопасно для горутин.
type Store struct {
	// ваши поля
}

func NewStore() *Store {
	// ваш код
	return &Store{}
}

func (s *Store) Get(k string) (string, bool) {
	// ваш код
	return "", false
}

// Set записывает значение. Это одно изменение в истории.
func (s *Store) Set(k, v string) {
	// ваш код
}

// Delete удаляет ключ. Удаление отсутствующего ключа ничего не меняет и в
// историю не попадает.
func (s *Store) Delete(k string) {
	// ваш код
}

// Batch выполняет fn; все изменения внутри отменяются и повторяются одним
// Undo/Redo. Вложенный Batch сливается с внешним. Пустой Batch в историю не
// попадает. Если fn паникует, уже сделанные изменения остаются одной
// записью в истории, а паника летит дальше.
func (s *Store) Batch(fn func()) {
	// ваш код
}

// Undo отменяет последнее изменение (или Batch). false — отменять нечего.
func (s *Store) Undo() bool {
	// ваш код
	return false
}

// Redo повторяет последнее отменённое. Любое новое изменение очищает redo.
func (s *Store) Redo() bool {
	// ваш код
	return false
}
