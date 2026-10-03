package main

// IsPaired сообщает, что все скобки в строке — (), [] и {} — сбалансированы
// и правильно вложены. Остальные символы не важны.
func IsPaired(value string) bool {
	pair := map[rune]rune{')': '(', ']': '[', '}': '{'}
	var stack []rune
	for _, r := range value {
		switch r {
		case '(', '[', '{':
			stack = append(stack, r)
		case ')', ']', '}':
			// Закрывающая обязана закрыть последнюю открытую — и именно свою.
			if len(stack) == 0 || stack[len(stack)-1] != pair[r] {
				return false
			}
			stack = stack[:len(stack)-1]
		}
	}
	// Незакрытые скобки — тоже ошибка.
	return len(stack) == 0
}
