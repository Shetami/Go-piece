package main

import "time"

// NewSampler — сэмплирование логов по ключу (как в zap).
// У каждого ключа своё окно длиной tick: оно начинается с первого
// сообщения ключа, а сообщение в момент start+tick или позже открывает
// новое окно со сброшенным счётчиком.
// В окне пропускаются первые first сообщений, дальше — каждое
// thereafter-е (first+thereafter, first+2·thereafter, …);
// thereafter <= 0 — после первых first больше ни одного.
// allow безопасна для вызова из многих горутин.
func NewSampler(first, thereafter int, tick time.Duration) (allow func(key string, now time.Time) bool) {
	// ваш код
	return func(key string, now time.Time) bool { return true }
}
