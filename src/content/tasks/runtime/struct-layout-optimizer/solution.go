package main

import "slices"

// Field — поле структуры: имя, размер и выравнивание в байтах
// (как unsafe.Sizeof и unsafe.Alignof). Size кратен Align, Align —
// степень двойки.
type Field struct {
	Name  string
	Size  uintptr
	Align uintptr
}

func alignUp(n, a uintptr) uintptr { return (n + a - 1) &^ (a - 1) }

// Layout раскладывает поля в заданном порядке по правилам компилятора gc
// и возвращает смещение каждого поля и размер всей структуры — ровно
// то, что показали бы unsafe.Offsetof и unsafe.Sizeof.
func Layout(fields []Field) (offsets []uintptr, size uintptr) {
	offsets = make([]uintptr, len(fields))
	var off uintptr
	maxAlign := uintptr(1)
	for i, f := range fields {
		off = alignUp(off, f.Align)
		offsets[i] = off
		off += f.Size
		maxAlign = max(maxAlign, f.Align)
	}
	// Правило gc: если последнее поле нулевого размера, добавляется байт —
	// иначе &s.last указывал бы за конец объекта, на соседний.
	if n := len(fields); n > 0 && fields[n-1].Size == 0 && off > 0 {
		off++
	}
	// Хвостовое выравнивание: в массиве каждый элемент должен быть выровнен.
	return offsets, alignUp(off, maxAlign)
}

// Optimize возвращает новый порядок тех же полей, при котором размер
// структуры минимален. Среди равноценных полей сохраняется исходный
// порядок. Входной слайс не меняется.
func Optimize(fields []Field) []Field {
	out := slices.Clone(fields)
	slices.SortStableFunc(out, func(a, b Field) int {
		// Поля нулевого размера — в начало: в конце они стоят лишний байт
		// плюс выравнивание.
		if az, bz := a.Size == 0, b.Size == 0; az != bz {
			if az {
				return -1
			}
			return 1
		}
		// Дальше по убыванию выравнивания: каждое поле попадает на
		// выровненное смещение без дыр.
		switch {
		case a.Align > b.Align:
			return -1
		case a.Align < b.Align:
			return 1
		}
		return 0
	})
	return out
}
