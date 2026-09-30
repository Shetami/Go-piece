package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// FormatRanges сворачивает номера в строку вида "1-3,5,7-8": по возрастанию,
// без повторов, два и больше подряд идущих числа — диапазоном "a-b".
// Вход в любом порядке, с повторами и отрицательными числами ("-3--1");
// вход не меняется. Пустой вход — "".
func FormatRanges(ids []int) string {
	s := slices.Clone(ids) // сортировать чужой слайс нельзя
	slices.Sort(s)
	s = slices.Compact(s)
	var b strings.Builder
	for i := 0; i < len(s); {
		j := i
		for j+1 < len(s) && s[j+1] == s[j]+1 {
			j++
		}
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(s[i]))
		if j > i {
			b.WriteByte('-')
			b.WriteString(strconv.Itoa(s[j]))
		}
		i = j + 1
	}
	return b.String()
}

// ParseRanges — обратное преобразование: "1-3,5" → [1 2 3 5], числа в том
// порядке, в каком записаны. "" — пустой результат без ошибки.
// Ошибка: пустой элемент ("1,,2"), не число, диапазон a-b с a > b.
func ParseRanges(s string) ([]int, error) {
	if s == "" {
		return nil, nil
	}
	var out []int
	for _, part := range strings.Split(s, ",") {
		if part == "" {
			return nil, fmt.Errorf("пустой элемент в %q", s)
		}
		// Разделитель диапазона ищем со второго символа: первый минус — знак числа.
		lo, hi := part, part
		if k := strings.IndexByte(part[1:], '-'); k >= 0 {
			lo, hi = part[:k+1], part[k+2:]
		}
		a, err := strconv.Atoi(lo)
		if err != nil {
			return nil, fmt.Errorf("элемент %q: %w", part, err)
		}
		b, err := strconv.Atoi(hi)
		if err != nil {
			return nil, fmt.Errorf("элемент %q: %w", part, err)
		}
		if a > b {
			return nil, fmt.Errorf("диапазон %q: начало больше конца", part)
		}
		for v := a; v <= b; v++ {
			out = append(out, v)
		}
	}
	return out, nil
}
