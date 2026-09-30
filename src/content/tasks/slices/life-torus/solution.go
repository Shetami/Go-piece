package main

// Life — поле игры «Жизнь» h×w на торе: края склеены, и соседи клетки у
// верхнего края лежат в нижней строке, у левого — в правом столбце.
// Поля — на ваше усмотрение.
type Life struct {
	h, w      int
	cur, next [][]bool // два поля: читаем из cur, пишем в next, потом меняем местами
}

// newGrid выделяет h строк одним общим массивом: две аллокации вместо h+1.
func newGrid(h, w int) [][]bool {
	cells := make([]bool, h*w)
	g := make([][]bool, h)
	for r := range g {
		g[r] = cells[r*w : (r+1)*w : (r+1)*w]
	}
	return g
}

// NewLife создаёт пустое поле. h < 3 или w < 3 — паника.
func NewLife(h, w int) *Life {
	if h < 3 || w < 3 {
		panic("NewLife: поле должно быть не меньше 3×3")
	}
	return &Life{h: h, w: w, cur: newGrid(h, w), next: newGrid(h, w)}
}

// Set делает клетку (r, c) живой или мёртвой.
func (l *Life) Set(r, c int, alive bool) { l.cur[r][c] = alive }

// Alive сообщает, жива ли клетка (r, c).
func (l *Life) Alive(r, c int) bool { return l.cur[r][c] }

// Step делает один шаг: живая клетка с 2 или 3 живыми соседями выживает,
// мёртвая ровно с 3 — оживает, остальные умирают. Все клетки меняются
// одновременно — по состоянию до шага. Step не выделяет память.
func (l *Life) Step() {
	for r := range l.h {
		up, down := (r-1+l.h)%l.h, (r+1)%l.h // +h: в Go -1 % h == -1
		for c := range l.w {
			left, right := (c-1+l.w)%l.w, (c+1)%l.w
			n := 0
			for _, rr := range [3]int{up, r, down} {
				for _, cc := range [3]int{left, c, right} {
					if (rr != r || cc != c) && l.cur[rr][cc] {
						n++
					}
				}
			}
			l.next[r][c] = n == 3 || (n == 2 && l.cur[r][c])
		}
	}
	l.cur, l.next = l.next, l.cur // обмен заголовков, без копирования клеток
}
