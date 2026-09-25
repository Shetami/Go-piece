package main

import "strings"

// Interner хранит по одному экземпляру каждой строки.
// Нулевое значение готово к работе. Не обязан быть потокобезопасным.
type Interner struct {
	m map[string]string
}

// Intern возвращает строку, равную s. Для равных строк всегда возвращается
// одна и та же строка — с общими байтами в памяти. Возвращаемая строка не
// должна удерживать в памяти то, частью чего была s (например, большой буфер).
func (in *Interner) Intern(s string) string {
	if v, ok := in.m[s]; ok {
		return v
	}
	if in.m == nil {
		in.m = make(map[string]string)
	}
	// Копия: s может быть подстрокой огромной строки, и хранить её как есть —
	// значит держать в памяти всю исходную строку навсегда.
	c := strings.Clone(s)
	in.m[c] = c
	return c
}

// Len — сколько разных строк сохранено.
func (in *Interner) Len() int { return len(in.m) }
