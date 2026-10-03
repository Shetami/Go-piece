package main

import "strings"

// Все «цифры», включая вычитательные пары, от больших к меньшим.
var romans = []struct {
	value int
	digit string
}{
	{1000, "M"}, {900, "CM"}, {500, "D"}, {400, "CD"},
	{100, "C"}, {90, "XC"}, {50, "L"}, {40, "XL"},
	{10, "X"}, {9, "IX"}, {5, "V"}, {4, "IV"}, {1, "I"},
}

// ToRoman записывает число от 1 до 3999 римскими цифрами.
func ToRoman(number int) string {
	var b strings.Builder
	for _, r := range romans {
		for number >= r.value {
			b.WriteString(r.digit)
			number -= r.value
		}
	}
	return b.String()
}
