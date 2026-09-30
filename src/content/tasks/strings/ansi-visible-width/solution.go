package main

import (
	"strings"
	"unicode/utf8"
)

// Строки для терминала содержат управляющие последовательности:
//
//   - CSI: ESC '[' , затем байты 0x30–0x3F (параметры), затем 0x20–0x2F
//     (промежуточные), затем один финальный байт 0x40–0x7E.
//     Примеры: "\x1b[31m" (красный), "\x1b[1;4m", "\x1b[0m", "\x1b[2K";
//   - OSC: ESC ']' … до BEL (0x07) или до ST (ESC '\'), например
//     гиперссылка "\x1b]8;;https://go.dev\x1b\\go.dev\x1b]8;;\x1b\\";
//   - ESC, за которым идёт любой другой символ, — двухсимвольная
//     последовательность (ESC + этот символ);
//   - оборванная последовательность в конце строки (например "\x1b[31")
//     невидима до конца строки.
//
// Видимый символ — любая руна вне последовательностей; ширина каждой — 1.

// reset сбрасывает все стили терминала.
const reset = "\x1b[0m"

// seqEnd возвращает индекс байта сразу за управляющей последовательностью,
// которая начинается в s[i] == ESC.
func seqEnd(s string, i int) int {
	if i+1 >= len(s) {
		return len(s)
	}
	switch s[i+1] {
	case '[':
		j := i + 2
		for j < len(s) && s[j] >= 0x20 && s[j] <= 0x3F { // параметры и промежуточные
			j++
		}
		if j < len(s) && s[j] >= 0x40 && s[j] <= 0x7E {
			return j + 1
		}
		return len(s) // оборвана или испорчена — до конца
	case ']':
		for j := i + 2; j < len(s); j++ {
			if s[j] == 0x07 {
				return j + 1
			}
			if s[j] == 0x1b && j+1 < len(s) && s[j+1] == '\\' {
				return j + 2
			}
		}
		return len(s)
	default:
		_, size := utf8.DecodeRuneInString(s[i+1:])
		return i + 1 + size
	}
}

// VisibleWidth — число видимых символов в s.
func VisibleWidth(s string) int {
	n := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b {
			i = seqEnd(s, i)
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
		n++
	}
	return n
}

// PadRight дополняет s пробелами справа до видимой ширины width.
// Строка шире width возвращается без изменений.
func PadRight(s string, width int) string {
	if w := VisibleWidth(s); w < width {
		return s + strings.Repeat(" ", width-w)
	}
	return s
}

// TruncateVisible оставляет в s не больше width видимых символов.
//
//   - все управляющие последовательности до места обрезки сохраняются,
//     включая идущие сразу за последним оставленным символом;
//   - последовательности после места обрезки выбрасываются;
//   - если строка обрезана и в оставленной части есть хотя бы одна
//     управляющая последовательность, в конец добавляется "\x1b[0m",
//     чтобы стиль не протёк дальше; если не обрезана — s без изменений.
func TruncateVisible(s string, width int) string {
	if VisibleWidth(s) <= width {
		return s
	}
	var b strings.Builder
	n, hasSeq := 0, false
	for i := 0; i < len(s); {
		if s[i] == 0x1b {
			j := seqEnd(s, i)
			b.WriteString(s[i:j])
			hasSeq = true
			i = j
			continue
		}
		if n == width {
			break // лимит набран: дальше только последовательности до следующего символа
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		b.WriteString(s[i : i+size])
		i += size
		n++
	}
	if hasSeq {
		b.WriteString(reset)
	}
	return b.String()
}
