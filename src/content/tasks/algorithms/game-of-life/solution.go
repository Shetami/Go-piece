package main

// Tick делает один шаг «Жизни» Конвея: 1 — живая клетка, 0 — мёртвая.
// Поле за краями считается мёртвым. Входная матрица не меняется.
func Tick(matrix [][]int) [][]int {
	next := make([][]int, len(matrix))
	for r, row := range matrix {
		next[r] = make([]int, len(row))
		for c, cell := range row {
			n := liveNeighbours(matrix, r, c)
			// Выживает при 2 или 3 соседях, рождается при ровно 3.
			if n == 3 || (cell == 1 && n == 2) {
				next[r][c] = 1
			}
		}
	}
	return next
}

func liveNeighbours(m [][]int, r, c int) int {
	n := 0
	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			rr, cc := r+dr, c+dc
			if (dr != 0 || dc != 0) && rr >= 0 && rr < len(m) && cc >= 0 && cc < len(m[rr]) {
				n += m[rr][cc]
			}
		}
	}
	return n
}
