package main

// Annotate проставляет в каждую пустую клетку поля число цветов (*)
// вокруг неё, считая диагонали. Клетку без цветов рядом оставляет пробелом.
func Annotate(garden []string) []string {
	out := make([]string, len(garden))
	for r, row := range garden {
		line := []byte(row)
		for c := range line {
			if line[c] == '*' {
				continue
			}
			n := 0
			for rr := r - 1; rr <= r+1; rr++ {
				for cc := c - 1; cc <= c+1; cc++ {
					if rr >= 0 && rr < len(garden) && cc >= 0 && cc < len(garden[rr]) && garden[rr][cc] == '*' {
						n++
					}
				}
			}
			if n > 0 {
				line[c] = byte('0' + n)
			}
		}
		out[r] = string(line)
	}
	return out
}
