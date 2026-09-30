package main

import (
	"errors"
	"fmt"
	"maps"
	"sync"
)

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
	get func() (map[string]string, error)
}

func NewDirectory(load func() (map[string]string, error)) *Directory {
	return &Directory{
		// OnceValues сам запоминает и значение, и ошибку.
		get: sync.OnceValues(func() (map[string]string, error) {
			m, err := load()
			if err != nil {
				return nil, fmt.Errorf("загрузка справочника: %w", err)
			}
			// Своя копия: загрузчик мог оставить мапу у себя и поменять.
			return maps.Clone(m), nil
		}),
	}
}

// Lookup возвращает название по коду или ErrUnknownCode.
func (d *Directory) Lookup(code string) (string, error) {
	m, err := d.get()
	if err != nil {
		return "", err
	}
	// После OnceValues мапа только читается — мьютекс не нужен.
	name, ok := m[code]
	if !ok {
		return "", fmt.Errorf("%q: %w", code, ErrUnknownCode)
	}
	return name, nil
}
