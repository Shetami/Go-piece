package main

// Balanced проверяет, что скобки (), [] и {} в строке сбалансированы
// и правильно вложены. Остальные символы не важны.
func Balanced(s string) bool {
	pairs := map[rune]rune{')': '(', ']': '[', '}': '{'}
	// Стек открытых скобок: закрывающая должна совпасть с последней открытой.
	var stack []rune
	for _, r := range s {
		switch r {
		case '(', '[', '{':
			stack = append(stack, r)
		case ')', ']', '}':
			if len(stack) == 0 || stack[len(stack)-1] != pairs[r] {
				return false
			}
			stack = stack[:len(stack)-1]
		}
	}
	// Незакрытые скобки — тоже дисбаланс.
	return len(stack) == 0
}
