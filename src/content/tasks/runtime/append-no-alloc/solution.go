package main

import "strconv"

// AppendPoint дописывает к dst точку в виде "(x, y)" и возвращает результат —
// как strconv.AppendInt. Если в dst хватает вместимости, память не выделяется.
func AppendPoint(dst []byte, x, y int) []byte {
	// Всё пишется прямо в dst: ни промежуточных строк, ни интерфейсов,
	// из-за которых fmt уводит аргументы в кучу.
	dst = append(dst, '(')
	dst = strconv.AppendInt(dst, int64(x), 10)
	dst = append(dst, ", "...)
	dst = strconv.AppendInt(dst, int64(y), 10)
	return append(dst, ')')
}
