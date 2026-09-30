package main

import (
	"errors"
	"slices"
)

var ErrReadOnly = errors.New("store is read-only")

// Store — хранилище строк.
type Store interface {
	Get(key string) (string, bool)
	Set(key, val string) error
	SetMany(kv map[string]string) error
	Delete(key string) error
	Keys() []string // по возрастанию
}

// MapStore — простая реализация Store (готова, менять не нужно).
type MapStore struct{ m map[string]string }

func NewMapStore() *MapStore { return &MapStore{m: map[string]string{}} }

func (s *MapStore) Get(k string) (string, bool) { v, ok := s.m[k]; return v, ok }
func (s *MapStore) Set(k, v string) error       { s.m[k] = v; return nil }
func (s *MapStore) Delete(k string) error       { delete(s.m, k); return nil }
func (s *MapStore) SetMany(kv map[string]string) error {
	for k, v := range kv {
		s.Set(k, v)
	}
	return nil
}
func (s *MapStore) Keys() []string {
	keys := make([]string, 0, len(s.m))
	for k := range s.m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// ReadOnly — вид на s только для чтения: Get и Keys как у s, любые
// изменения возвращают ErrReadOnly и s не трогают.
func ReadOnly(s Store) Store {
	// ваш код
	return s
}

// Prefixed — вид на s, где все ключи получают префикс prefix: Set("a", …)
// пишет в s ключ prefix+"a", Get/Delete/SetMany — так же. Keys возвращает
// только ключи s с этим префиксом, уже без него, по возрастанию.
// Обёртки комбинируются: Prefixed(Prefixed(s, "a/"), "b/") пишет в "a/b/…".
func Prefixed(s Store, prefix string) Store {
	// ваш код
	return s
}
