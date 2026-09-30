package main

import "errors"

// ErrUnknownCode — такого кода в справочнике нет.
var ErrUnknownCode = errors.New("unknown code")

// Directory — справочник «код → название» (страны, валюты), который
// грузится из базы при первом обращении, а не при старте сервиса.
//
// load вызывается не раньше первого Lookup и ровно один раз за всё время
// жизни Directory, сколько бы горутин ни звали Lookup одновременно.
// Если загрузка вернула ошибку, её получают все Lookup (errors.Is
// должен её находить) — повторно load не вызывается.
type Directory struct {
	// ваши поля
}

func NewDirectory(load func() (map[string]string, error)) *Directory {
	// ваш код
	return &Directory{}
}

// Lookup возвращает название по коду или ErrUnknownCode.
func (d *Directory) Lookup(code string) (string, error) {
	// ваш код
	return "", nil
}
