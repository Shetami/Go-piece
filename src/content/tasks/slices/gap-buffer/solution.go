package main

// Editor — текстовый буфер с курсором на gap buffer: символы до курсора
// лежат в начале buf, после курсора — в конце, а между ними «дыра».
// Нулевое значение — пустой буфер с курсором в начале.
type Editor struct {
	buf      []rune
	gapStart int // позиция курсора
	gapEnd   int // первый символ после курсора
}

// grow расширяет дыру хотя бы до need символов.
func (e *Editor) grow(need int) {
	if e.gapEnd-e.gapStart >= need {
		return
	}
	used := len(e.buf) - (e.gapEnd - e.gapStart)
	size := max(2*len(e.buf), used+need, 16) // удвоение — амортизированно O(1)
	nb := make([]rune, size)
	copy(nb, e.buf[:e.gapStart])
	// Хвост после курсора переезжает в самый конец нового массива.
	tail := len(e.buf) - e.gapEnd
	copy(nb[size-tail:], e.buf[e.gapEnd:])
	e.buf, e.gapEnd = nb, size-tail
}

// Insert вставляет s перед курсором; курсор оказывается после вставленного.
// Набор по одному символу — амортизированно O(1).
func (e *Editor) Insert(s string) {
	rs := []rune(s) // курсор считается в символах, а не в байтах
	e.grow(len(rs))
	e.gapStart += copy(e.buf[e.gapStart:], rs)
}

// Backspace удаляет до n символов перед курсором и возвращает, сколько удалил.
func (e *Editor) Backspace(n int) int {
	k := min(max(n, 0), e.gapStart)
	e.gapStart -= k // удалённое просто становится частью дыры
	return k
}

// Delete удаляет до n символов после курсора и возвращает, сколько удалил.
func (e *Editor) Delete(n int) int {
	k := min(max(n, 0), len(e.buf)-e.gapEnd)
	e.gapEnd += k
	return k
}

// Move сдвигает курсор на d символов (d < 0 — влево), не выходя за границы текста.
// Стоимость — O(|d|), а не O(длины текста).
func (e *Editor) Move(d int) {
	if d < 0 {
		k := min(-d, e.gapStart)
		// k символов перед дырой переезжают за неё. Области могут
		// перекрываться, если дыра меньше k, — copy это умеет.
		copy(e.buf[e.gapEnd-k:e.gapEnd], e.buf[e.gapStart-k:e.gapStart])
		e.gapStart -= k
		e.gapEnd -= k
	} else {
		k := min(d, len(e.buf)-e.gapEnd)
		copy(e.buf[e.gapStart:e.gapStart+k], e.buf[e.gapEnd:e.gapEnd+k])
		e.gapStart += k
		e.gapEnd += k
	}
}

// Cursor — позиция курсора в символах (рунах) от начала текста.
func (e *Editor) Cursor() int { return e.gapStart }

// String возвращает весь текст.
func (e *Editor) String() string {
	return string(e.buf[:e.gapStart]) + string(e.buf[e.gapEnd:])
}
