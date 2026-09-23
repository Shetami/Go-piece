package main

import (
	"fmt"
	"unsafe"
)

// toString превращает байты в строку без копирования — «для скорости».
func toString(b []byte) string {
	return unsafe.String(&b[0], len(b))
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
