package main

import (
	"io"
	"strings"
	"unicode/utf8"
)

// WriteTable печатает таблицу в w.
//
//   - rows[0] — заголовок, под ним строка из '-' шириной каждой колонки;
//   - ширина колонки — максимальная длина ячейки в символах (рунах) по всем
//     строкам, включая заголовок;
//   - колонки разделены двумя пробелами;
//   - строки бывают разной длины: недостающие ячейки считаются пустыми,
//     колонок столько, сколько в самой длинной строке;
//   - числа (необязательный '-', цифры, необязательно '.' и цифры) в
//     строках данных выравниваются по правому краю, всё остальное — и весь
//     заголовок — по левому;
//   - хвостовых пробелов в строках нет, каждая строка заканчивается "\n";
//   - пустой rows — ничего не пишется.
//
// Ошибка записи в w возвращается как есть.
func WriteTable(w io.Writer, rows [][]string) error {
	if len(rows) == 0 {
		return nil
	}
	ncol := 0
	for _, r := range rows {
		ncol = max(ncol, len(r))
	}
	widths := make([]int, ncol)
	for _, r := range rows {
		for i, c := range r {
			widths[i] = max(widths[i], utf8.RuneCountInString(c)) // символы, не байты
		}
	}
	var b strings.Builder
	line := func(cells []string, header bool) {
		var l strings.Builder
		for i := range ncol {
			if i > 0 {
				l.WriteString("  ")
			}
			c := ""
			if i < len(cells) {
				c = cells[i] // короткая строка — недостающие ячейки пустые
			}
			pad := strings.Repeat(" ", widths[i]-utf8.RuneCountInString(c))
			if !header && isNumber(c) {
				l.WriteString(pad + c)
			} else {
				l.WriteString(c + pad)
			}
		}
		// Последняя колонка или пустые ячейки в конце дают хвостовые пробелы.
		b.WriteString(strings.TrimRight(l.String(), " "))
		b.WriteByte('\n')
	}
	line(rows[0], true)
	dashes := make([]string, ncol)
	for i, wd := range widths {
		dashes[i] = strings.Repeat("-", wd)
	}
	b.WriteString(strings.Join(dashes, "  ") + "\n")
	for _, r := range rows[1:] {
		line(r, false)
	}
	// Одна запись вместо десятков мелких; ошибку писателя отдаём наверх.
	_, err := io.WriteString(w, b.String())
	return err
}

func isNumber(s string) bool {
	s = strings.TrimPrefix(s, "-")
	intPart, frac, hasDot := strings.Cut(s, ".")
	digits := func(x string) bool {
		if x == "" {
			return false
		}
		for i := 0; i < len(x); i++ {
			if x[i] < '0' || x[i] > '9' {
				return false
			}
		}
		return true
	}
	return digits(intPart) && (!hasDot || digits(frac))
}
