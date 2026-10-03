package main

// Factors раскладывает натуральное число на простые множители
// и возвращает их по возрастанию, с повторениями: 12 → [2 2 3].
func Factors(value int) []int {
	var out []int
	// Делители перебираем по возрастанию, поэтому каждый найденный — простой:
	// все меньшие простые из value уже вынесены.
	for d := 2; d*d <= value; d++ {
		for value%d == 0 {
			out = append(out, d)
			value /= d
		}
	}
	// Что осталось больше единицы — простой множитель больше корня.
	if value > 1 {
		out = append(out, value)
	}
	return out
}
