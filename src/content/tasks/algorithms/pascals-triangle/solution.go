package main

// Pascal возвращает первые count строк треугольника Паскаля.
func Pascal(count int) [][]int {
	rows := make([][]int, count)
	for i := range rows {
		row := make([]int, i+1)
		row[0], row[i] = 1, 1
		for j := 1; j < i; j++ {
			row[j] = rows[i-1][j-1] + rows[i-1][j]
		}
		rows[i] = row
	}
	return rows
}
