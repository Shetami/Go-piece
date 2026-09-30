package main

import (
	"sync"
	"time"
)

// NewSampler — сэмплирование логов по ключу (как в zap).
// У каждого ключа своё окно длиной tick: оно начинается с первого
// сообщения ключа, а сообщение в момент start+tick или позже открывает
// новое окно со сброшенным счётчиком.
// В окне пропускаются первые first сообщений, дальше — каждое
// thereafter-е (first+thereafter, first+2·thereafter, …);
// thereafter <= 0 — после первых first больше ни одного.
// allow безопасна для вызова из многих горутин.
func NewSampler(first, thereafter int, tick time.Duration) (allow func(key string, now time.Time) bool) {
	type window struct {
		start time.Time
		n     int
	}
	var (
		mu sync.Mutex
		ws = make(map[string]*window)
	)
	return func(key string, now time.Time) bool {
		mu.Lock()
		defer mu.Unlock()
		w := ws[key]
		if w == nil || !now.Before(w.start.Add(tick)) {
			w = &window{start: now}
			ws[key] = w
		}
		w.n++
		if w.n <= first {
			return true
		}
		// thereafter <= 0 проверяем до деления: % 0 — паника.
		return thereafter > 0 && (w.n-first)%thereafter == 0
	}
}
