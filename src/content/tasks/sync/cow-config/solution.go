package main

import (
	"maps"
	"sync/atomic"
)

// Config — конфигурация сервиса: лимиты запросов по клиентам.
type Config struct {
	Version int
	Limits  map[string]int
}

// clone — глубокая копия: копия структуры делит с оригиналом мапу.
func (c *Config) clone() *Config {
	cp := *c
	cp.Limits = maps.Clone(c.Limits)
	if cp.Limits == nil {
		cp.Limits = map[string]int{}
	}
	return &cp
}

// Holder хранит текущий конфиг. Load вызывают на каждый запрос, поэтому
// он не должен брать блокировок и никогда не ждёт Update. Конфиг,
// полученный из Load, — неизменяемый снимок: его нельзя менять, и сам
// Holder его тоже никогда не меняет.
type Holder struct {
	cur atomic.Pointer[Config]
}

// NewHolder публикует начальный конфиг. Вызывающий может потом менять
// свою копию initial — на Holder это не влияет.
func NewHolder(initial Config) *Holder {
	h := &Holder{}
	h.cur.Store(initial.clone())
	return h
}

// Load возвращает текущий снимок.
func (h *Holder) Load() *Config {
	return h.cur.Load()
}

// Update применяет fn к копии текущего конфига и публикует результат с
// Version на единицу больше. Конкурентные Update не теряют изменений
// друг друга. fn может быть вызвана больше одного раза, поэтому должна
// лишь менять переданную копию.
func (h *Holder) Update(fn func(c *Config)) {
	for {
		old := h.cur.Load()
		next := old.clone() // меняем копию, старый снимок могут читать прямо сейчас
		fn(next)
		next.Version = old.Version + 1
		// Публикуем, только если за время fn никто не успел опубликовать
		// своё; иначе повторяем поверх свежей версии.
		if h.cur.CompareAndSwap(old, next) {
			return
		}
	}
}
