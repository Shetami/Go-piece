package main

// AppendKV дописывает к dst запись вида key=val (для строк — key="val"
// с экранированием как у strconv.Quote) и возвращает результат.
// Если вместимости dst хватает, функция не выделяет память вообще.
//
//	AppendKV(nil, "port", 8080)    → port=8080
//	AppendKV(nil, "user", "аня")   → user="аня"
//	AppendKV(nil, "debug", true)   → debug=true
//
// val может быть int, string или bool; для остальных типов — key=?
func AppendKV(dst []byte, key string, val any) []byte {
	// ваш код
	return dst
}
