package main

// Spiral обходит прямоугольную матрицу по спирали: по часовой стрелке,
// начиная с левого верхнего угла, снаружи внутрь. Каждый элемент — ровно один раз.
// Пустая матрица (нет строк или строки пустые) — пустой результат.
// Строки разной длины — паника. Матрица не меняется.
// Под результат — не больше одной аллокации.
func Spiral[T any](m [][]T) []T {
	if len(m) == 0 {
		return nil
	}
	cols := len(m[0])
	for _, row := range m {
		if len(row) != cols {
			panic("Spiral: строки разной длины")
		}
	}
	out := make([]T, 0, len(m)*cols) // размер известен заранее
	top, bottom, left, right := 0, len(m)-1, 0, cols-1
	for top <= bottom && left <= right {
		for j := left; j <= right; j++ {
			out = append(out, m[top][j])
		}
		top++
		for i := top; i <= bottom; i++ {
			out = append(out, m[i][right])
		}
		right--
		// Осталась одна строка или один столбец — обратный проход повторил бы их.
		if top <= bottom {
			for j := right; j >= left; j-- {
				out = append(out, m[bottom][j])
			}
			bottom--
		}
		if left <= right {
			for i := bottom; i >= top; i-- {
				out = append(out, m[i][left])
			}
			left++
		}
	}
	return out
}
