package main

// LongestUnique возвращает длину (в символах) самой длинной подстроки
// без повторяющихся символов. Строка в UTF-8. Сложность — O(n).
func LongestUnique(s string) int {
	last := map[rune]int{} // руна → её последняя позиция (в рунах)
	best, start, pos := 0, 0, 0
	for _, r := range s {
		// Повтор внутри окна — сдвигаем левую границу за прошлое вхождение.
		// Вхождения левее окна не в счёт: start назад не двигается.
		if p, ok := last[r]; ok && p >= start {
			start = p + 1
		}
		last[r] = pos
		best = max(best, pos-start+1)
		pos++
	}
	return best
}
