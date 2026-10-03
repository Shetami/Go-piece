package main

// Spiral возвращает квадратную матрицу size×size, заполненную числами
// от 1 до size² по спирали: по часовой стрелке, начиная с левого верхнего угла.
func Spiral(size int) [][]int {
	m := make([][]int, size)
	for i := range m {
		m[i] = make([]int, size)
	}
	// Направления по часовой: вправо, вниз, влево, вверх.
	dr := [4]int{0, 1, 0, -1}
	dc := [4]int{1, 0, -1, 0}
	r, c, dir := 0, 0, 0
	for n := 1; n <= size*size; n++ {
		m[r][c] = n
		nr, nc := r+dr[dir], c+dc[dir]
		// Упёрлись в край или в уже заполненную клетку — поворот.
		if nr < 0 || nr >= size || nc < 0 || nc >= size || m[nr][nc] != 0 {
			dir = (dir + 1) % 4
			nr, nc = r+dr[dir], c+dc[dir]
		}
		r, c = nr, nc
	}
	return m
}
