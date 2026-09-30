package main

import (
	"fmt"
	"strings"
)

// SyntaxError — ошибка разбора командной строки; Pos — позиция в символах
// (рунах, с нуля).
type SyntaxError struct {
	Pos int
	Msg string
}

func (e *SyntaxError) Error() string { return fmt.Sprintf("символ %d: %s", e.Pos, e.Msg) }

// SplitArgs разбивает командную строку на аргументы, как POSIX shell, но
// без подстановок переменных, glob и прочего:
//
//	cp "мой файл.txt" 'a b' dir\ name → ["cp", "мой файл.txt", "a b", "dir name"]
//
// Правила:
//   - аргументы разделены пробелами, табами и переводами строк;
//   - '…' — всё внутри буквально, '\' внутри одинарных кавычек не особый;
//   - "…" — внутри '\' экранирует только '"' и '\'; перед любым другим
//     символом '\' остаётся как есть ("a\nb" → a\nb);
//   - вне кавычек '\' экранирует любой следующий символ, в том числе
//     пробел и кавычку;
//   - кавычки склеиваются с соседним текстом: a"b c"'d' → `ab cd`;
//   - пустые кавычки (двойные или одинарные) — пустой аргумент;
//   - незакрытая кавычка — *SyntaxError с Pos открывающей кавычки;
//     '\' в конце строки — *SyntaxError с Pos этой '\'.
//
// Пустая строка или строка из пробелов — пустой результат без ошибки.
func SplitArgs(line string) ([]string, error) {
	rs := []rune(line)
	var args []string
	var cur strings.Builder
	inArg := false // есть начатый аргумент, даже пустой (после "")
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		switch {
		case r == ' ' || r == '\t' || r == '\n':
			if inArg {
				args = append(args, cur.String())
				cur.Reset()
				inArg = false
			}
		case r == '\\':
			if i+1 == len(rs) {
				return nil, &SyntaxError{Pos: i, Msg: "'\\' в конце строки"}
			}
			i++
			cur.WriteRune(rs[i])
			inArg = true
		case r == '\'':
			open := i
			end := -1
			for j := i + 1; j < len(rs); j++ {
				if rs[j] == '\'' {
					end = j
					break
				}
			}
			if end < 0 {
				return nil, &SyntaxError{Pos: open, Msg: "незакрытая одинарная кавычка"}
			}
			cur.WriteString(string(rs[i+1 : end]))
			i = end
			inArg = true
		case r == '"':
			open := i
			i++
			for ; i < len(rs) && rs[i] != '"'; i++ {
				if rs[i] == '\\' && i+1 < len(rs) && (rs[i+1] == '"' || rs[i+1] == '\\') {
					i++
				}
				cur.WriteRune(rs[i])
			}
			if i == len(rs) {
				return nil, &SyntaxError{Pos: open, Msg: "незакрытая двойная кавычка"}
			}
			inArg = true
		default:
			cur.WriteRune(r)
			inArg = true
		}
	}
	if inArg {
		args = append(args, cur.String())
	}
	return args, nil
}
