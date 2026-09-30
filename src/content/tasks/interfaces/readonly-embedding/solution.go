package main

import (
	"errors"
	"slices"
	"strings"
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

// readOnly встраивает Store: Get и Keys достаются как есть,
// а все изменяющие методы переопределены. Все — включая SetMany.
type readOnly struct{ Store }

func (readOnly) Set(string, string) error        { return ErrReadOnly }
func (readOnly) SetMany(map[string]string) error { return ErrReadOnly }
func (readOnly) Delete(string) error             { return ErrReadOnly }

// ReadOnly — вид на s только для чтения: Get и Keys как у s, любые
// изменения возвращают ErrReadOnly и s не трогают.
func ReadOnly(s Store) Store { return readOnly{s} }

// prefixed переопределяет все методы: встраивание здесь не помогло бы,
// любой забытый метод работал бы с ключами без префикса.
type prefixed struct {
	s      Store
	prefix string
}

func (p prefixed) Get(k string) (string, bool) { return p.s.Get(p.prefix + k) }
func (p prefixed) Set(k, v string) error       { return p.s.Set(p.prefix+k, v) }
func (p prefixed) Delete(k string) error       { return p.s.Delete(p.prefix + k) }

func (p prefixed) SetMany(kv map[string]string) error {
	pk := make(map[string]string, len(kv)) // чужую мапу не меняем
	for k, v := range kv {
		pk[p.prefix+k] = v
	}
	return p.s.SetMany(pk)
}

func (p prefixed) Keys() []string {
	var out []string
	for _, k := range p.s.Keys() {
		if rest, ok := strings.CutPrefix(k, p.prefix); ok {
			out = append(out, rest)
		}
	}
	return out // порядок сохраняется: у всех ключей одинаковый префикс
}

// Prefixed — вид на s, где все ключи получают префикс prefix: Set("a", …)
// пишет в s ключ prefix+"a", Get/Delete/SetMany — так же. Keys возвращает
// только ключи s с этим префиксом, уже без него, по возрастанию.
// Обёртки комбинируются: Prefixed(Prefixed(s, "a/"), "b/") пишет в "a/b/…".
func Prefixed(s Store, prefix string) Store { return prefixed{s, prefix} }
