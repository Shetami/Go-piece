package main

import "fmt"

// toString превращает байты в строку. Обычная конверсия копирует данные —
// и только поэтому строка остаётся неизменной, когда буфер переиспользуют.
func toString(b []byte) string {
	return string(b)
}

func main() {
	buf := []byte("привет")
	s := toString(buf)

	seen := map[string]int{s: 1}

	copy(buf, "пока!!") // буфер переиспользуется под следующее сообщение

	fmt.Println(s)
	_, ok := seen["привет"]
	fmt.Println(ok)
}
