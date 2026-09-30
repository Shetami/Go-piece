package main

// Config — конфигурация сервиса: лимиты запросов по клиентам.
type Config struct {
	Version int
	Limits  map[string]int
}

// Holder хранит текущий конфиг. Load вызывают на каждый запрос, поэтому
// он не должен брать блокировок и никогда не ждёт Update. Конфиг,
// полученный из Load, — неизменяемый снимок: его нельзя менять, и сам
// Holder его тоже никогда не меняет.
type Holder struct {
	// ваши поля
}

// NewHolder публикует начальный конфиг. Вызывающий может потом менять
// свою копию initial — на Holder это не влияет.
func NewHolder(initial Config) *Holder {
	// ваш код
	return &Holder{}
}

// Load возвращает текущий снимок.
func (h *Holder) Load() *Config {
	// ваш код
	return &Config{}
}

// Update применяет fn к копии текущего конфига и публикует результат с
// Version на единицу больше. Конкурентные Update не теряют изменений
// друг друга. fn может быть вызвана больше одного раза, поэтому должна
// лишь менять переданную копию.
func (h *Holder) Update(fn func(c *Config)) {
	// ваш код
}
