package main

import (
	"strings"
	"unicode"
)

// Words возвращает функцию-итератор по словам строки s.
// Каждый вызов возвращает следующее слово и true; когда слова кончились —
// "" и false (и дальше всегда "" и false).
// Слова разделяются любыми пробельными символами.
// Строку заранее целиком не резать: каждый вызов находит одно следующее слово.
func Words(s string) func() (string, bool) {
	// Состояние итератора — непрочитанный остаток строки.
	rest := s
	return func() (string, bool) {
		rest = strings.TrimLeftFunc(rest, unicode.IsSpace)
		if rest == "" {
			return "", false
		}
		end := strings.IndexFunc(rest, unicode.IsSpace)
		if end < 0 {
			end = len(rest)
		}
		w := rest[:end]
		rest = rest[end:]
		return w, true
	}
}
