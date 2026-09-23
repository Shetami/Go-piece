package main

import "fmt"

// AppendPoint дописывает к dst точку в виде "(x, y)" и возвращает результат —
// как strconv.AppendInt. Если в dst хватает вместимости, память не выделяется.
func AppendPoint(dst []byte, x, y int) []byte {
	// ваш код
	return append(dst, fmt.Sprintf("(%d, %d)", x, y)...)
}
