package main

// FirstUnique возвращает первый символ строки, который встречается в ней
// ровно один раз. Строка — в UTF-8, считать надо символы (руны), а не байты.
// Если такого символа нет — false.
func FirstUnique(s string) (rune, bool) {
	counts := make(map[rune]int)
	for _, r := range s {
		counts[r]++
	}
	// Второй проход — по строке, а не по мапе: у мапы нет порядка.
	for _, r := range s {
		if counts[r] == 1 {
			return r, true
		}
	}
	return 0, false
}
