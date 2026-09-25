package main

import "strconv"

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
	dst = append(dst, key...)
	dst = append(dst, '=')
	// strconv.Append* пишут прямо в dst — без промежуточных строк,
	// которые дали бы Itoa, Quote или fmt.Sprintf.
	switch v := val.(type) {
	case int:
		return strconv.AppendInt(dst, int64(v), 10)
	case string:
		return strconv.AppendQuote(dst, v)
	case bool:
		return strconv.AppendBool(dst, v)
	default:
		return append(dst, '?')
	}
}
