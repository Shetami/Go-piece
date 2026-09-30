package main

import "iter"

// Trie — префиксное дерево со строковыми ключами. Нулевое значение
// готово к работе.
type Trie[V any] struct {
	// ваши поля
}

// Put записывает значение ключа; существующий ключ перезаписывается.
// Пустая строка — обычный ключ.
func (t *Trie[V]) Put(key string, v V) {
	// ваш код
}

// Get возвращает значение ключа и true, если ключ есть.
func (t *Trie[V]) Get(key string) (V, bool) {
	// ваш код
	var zero V
	return zero, false
}

// Len возвращает число различных ключей.
func (t *Trie[V]) Len() int {
	// ваш код
	return 0
}

// WithPrefix обходит все ключи, начинающиеся с prefix (включая сам
// prefix, если он ключ), в порядке возрастания — как sort.Strings.
func (t *Trie[V]) WithPrefix(prefix string) iter.Seq2[string, V] {
	// ваш код
	return func(yield func(string, V) bool) {}
}
