package main

import "errors"

var ErrClosed = errors.New("resource already closed")

// Tracker выдаёт ресурсы (условные соединения) и ловит забытые Close:
// ресурс, ставший недостижимым без Close, после сборки мусора попадает
// в список утёкших — так os.File и net.Conn закрывают за собой
// дескрипторы. Безопасен для конкурентного использования.
// Сам Tracker не должен мешать сборщику мусора собирать ресурсы.
type Tracker struct {
	// ваши поля
}

// Resource — выданный ресурс.
type Resource struct {
	// ваши поля
}

// Open выдаёт новый ресурс с именем name.
func (t *Tracker) Open(name string) *Resource {
	// ваш код
	return &Resource{}
}

// Close закрывает ресурс; повторный Close возвращает ErrClosed.
// Закрытый ресурс утёкшим не считается.
func (r *Resource) Close() error {
	// ваш код
	return nil
}

// Leaked — имена утёкших ресурсов (собранных без Close), по возрастанию.
func (t *Tracker) Leaked() []string {
	// ваш код
	return nil
}

// OpenCount — сколько ресурсов открыто сейчас: выданы, не закрыты и не утекли.
func (t *Tracker) OpenCount() int {
	// ваш код
	return 0
}
