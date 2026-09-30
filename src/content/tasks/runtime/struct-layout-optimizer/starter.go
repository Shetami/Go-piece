package main

// Field — поле структуры: имя, размер и выравнивание в байтах
// (как unsafe.Sizeof и unsafe.Alignof). Size кратен Align, Align —
// степень двойки.
type Field struct {
	Name  string
	Size  uintptr
	Align uintptr
}

// Layout раскладывает поля в заданном порядке по правилам компилятора gc
// и возвращает смещение каждого поля и размер всей структуры — ровно
// то, что показали бы unsafe.Offsetof и unsafe.Sizeof.
func Layout(fields []Field) (offsets []uintptr, size uintptr) {
	// ваш код
	return nil, 0
}

// Optimize возвращает новый порядок тех же полей, при котором размер
// структуры минимален. Среди равноценных полей сохраняется исходный
// порядок. Входной слайс не меняется.
func Optimize(fields []Field) []Field {
	// ваш код
	return fields
}
