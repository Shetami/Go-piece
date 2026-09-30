package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const ellipsis = "…" // 3 байта в UTF-8

// TruncateBytes обрезает s так, чтобы результат занимал не больше limit
// байт — например, под колонку VARCHAR(n) в байтах или лимит push-уведомления.
//
//   - если s помещается в limit — возвращается без изменений;
//   - иначе результат — префикс s плюс "…", всё вместе не длиннее limit;
//   - префикс состоит из целых «кластеров»: руна не разрывается, и от
//     символа не отрываются идущие за ним комбинирующие знаки
//     (unicode.Mn и unicode.Me, например U+0306 в «и» + «̆» = «й»),
//     вариационные селекторы U+FE00–U+FE0F и модификаторы тона кожи
//     U+1F3FB–U+1F3FF; ZWJ (U+200D) приклеивает к кластеру и себя, и
//     следующую руну (👨‍👩‍👧 — один кластер);
//   - пробельные символы в конце префикса убираются перед "…";
//   - если не помещается даже "…" (limit < 3) — пустая строка;
//   - байт невалидного UTF-8 — отдельный кластер длиной 1 байт.
func TruncateBytes(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	budget := limit - len(ellipsis)
	if budget < 0 {
		return ""
	}
	end := 0 // граница последнего целого кластера, который помещается
	for i := 0; i < len(s); {
		next := clusterEnd(s, i)
		if next > budget {
			break
		}
		end, i = next, next
	}
	return strings.TrimRightFunc(s[:end], unicode.IsSpace) + ellipsis
}

// clusterEnd возвращает байтовую границу кластера, который начинается в i.
func clusterEnd(s string, i int) int {
	_, size := utf8.DecodeRuneInString(s[i:]) // невалидный байт: size == 1
	i += size
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == '‍':
			// ZWJ склеивает с кластером следующую руну.
			i += size
			if i < len(s) {
				_, sz := utf8.DecodeRuneInString(s[i:])
				i += sz
			}
		case isExtend(r):
			i += size
		default:
			return i
		}
	}
	return i
}

func isExtend(r rune) bool {
	return unicode.In(r, unicode.Mn, unicode.Me) ||
		(r >= 0xFE00 && r <= 0xFE0F) ||
		(r >= 0x1F3FB && r <= 0x1F3FF)
}
