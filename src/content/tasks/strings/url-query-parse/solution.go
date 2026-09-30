package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// QueryError — ошибка разбора. Pos — байтовое смещение в исходной строке.
type QueryError struct {
	Pos int
	Msg string
}

func (e *QueryError) Error() string { return fmt.Sprintf("query, байт %d: %s", e.Pos, e.Msg) }

// ParseQuery разбирает query-строку URL без net/url:
// "q=%D0%B3%D0%BE+lang&tag=a&tag=b&debug" →
// {"q": ["го lang"], "tag": ["a", "b"], "debug": [""]}.
//
//   - пары разделены '&' (';' — обычный символ); пустые пары пропускаются;
//   - ключ и значение разделены первым '='; остальные '=' — часть значения;
//     пара без '=' — ключ с пустым значением;
//   - и в ключе, и в значении '+' означает пробел, а %XX — байт с кодом XX
//     (цифры в любом регистре); "%2B" — буквальный '+', "%26" — '&';
//   - плохая последовательность ('%' без двух hex-цифр) — *QueryError,
//     Pos — байт знака '%' в исходной строке;
//   - после декодирования ключ и значение должны быть валидным UTF-8,
//     иначе *QueryError с Pos первого байта этого ключа или значения;
//   - значения одного ключа — в порядке появления.
//
// При ошибке результат — nil.
func ParseQuery(q string) (map[string][]string, error) {
	res := map[string][]string{}
	pos := 0 // смещение текущей пары в q
	for _, pair := range strings.Split(q, "&") {
		start := pos
		pos += len(pair) + 1
		if pair == "" {
			continue
		}
		rawKey, rawVal, _ := strings.Cut(pair, "=")
		key, err := unescape(rawKey, start)
		if err != nil {
			return nil, err
		}
		valStart := start + len(rawKey) + 1
		val, err := unescape(rawVal, valStart)
		if err != nil {
			return nil, err
		}
		res[key] = append(res[key], val)
	}
	return res, nil
}

// unescape декодирует '+' и %XX; base — смещение s в исходной строке.
func unescape(s string, base int) (string, error) {
	if !strings.ContainsAny(s, "%+") {
		if !utf8.ValidString(s) {
			return "", &QueryError{Pos: base, Msg: "невалидный UTF-8"}
		}
		return s, nil
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '+':
			b.WriteByte(' ') // '+' смотрим в сыром виде: "%2B" пробелом не станет
		case '%':
			if i+2 >= len(s) {
				return "", &QueryError{Pos: base + i, Msg: "обрезанная %-последовательность"}
			}
			hi, ok1 := unhex(s[i+1])
			lo, ok2 := unhex(s[i+2])
			if !ok1 || !ok2 {
				return "", &QueryError{Pos: base + i, Msg: "некорректная %-последовательность"}
			}
			b.WriteByte(hi<<4 | lo)
			i += 2
		default:
			b.WriteByte(c)
		}
	}
	out := b.String()
	// Кириллица приходит байтами по отдельности: "%D0" без пары — битый UTF-8.
	if !utf8.ValidString(out) {
		return "", &QueryError{Pos: base, Msg: "невалидный UTF-8 после декодирования"}
	}
	return out, nil
}

func unhex(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}
