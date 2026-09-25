package main

// Transpose возвращает транспонированную матрицу: строки становятся столбцами.
// Все строки m одной длины. Пустая матрица → пустая (nil допустим).
// Исходную матрицу не менять.
func Transpose(m [][]int) [][]int {
	if len(m) == 0 {
		return nil
	}
	rows, cols := len(m), len(m[0])
	// Один массив на все клетки — одно выделение вместо cols.
	cells := make([]int, rows*cols)
	out := make([][]int, cols)
	for c := range out {
		out[c] = cells[c*rows : (c+1)*rows : (c+1)*rows]
		for r := range rows {
			out[c][r] = m[r][c]
		}
	}
	return out
}
