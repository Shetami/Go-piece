package main

import "unicode"

// IsPalindrome сообщает, читается ли фраза одинаково в обе стороны.
// Регистр не важен, учитываются только буквы и цифры.
func IsPalindrome(s string) bool {
	// Собираем только значимые руны в одном регистре — байты тут не годятся:
	// кириллица занимает по два байта.
	var rs []rune
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			rs = append(rs, unicode.ToLower(r))
		}
	}
	for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
		if rs[i] != rs[j] {
			return false
		}
	}
	return true
}
