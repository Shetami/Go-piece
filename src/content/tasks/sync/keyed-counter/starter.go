package main

// Counter — счётчики по ключам, безопасные для использования из многих горутин.
type Counter struct {
	// ваши поля
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) Inc(key string) {
	// ваш код
}

func (c *Counter) Get(key string) int {
	// ваш код
	return 0
}

// Snapshot возвращает копию всех счётчиков. Изменения копии
// не должны влиять на Counter, и наоборот.
func (c *Counter) Snapshot() map[string]int {
	// ваш код
	return nil
}
