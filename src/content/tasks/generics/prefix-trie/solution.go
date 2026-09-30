package main

import (
	"iter"
	"maps"
	"slices"
)

type trieNode[V any] struct {
	children map[byte]*trieNode[V]
	has      bool // есть ли ключ, заканчивающийся здесь: нулевое V — тоже значение
	val      V
}

// Trie — префиксное дерево со строковыми ключами. Нулевое значение
// готово к работе.
type Trie[V any] struct {
	root trieNode[V]
	n    int
}

// Put записывает значение ключа; существующий ключ перезаписывается.
// Пустая строка — обычный ключ.
func (t *Trie[V]) Put(key string, v V) {
	nd := &t.root
	for i := 0; i < len(key); i++ { // по байтам: порядок байтов UTF-8 = порядок строк
		if nd.children == nil {
			nd.children = make(map[byte]*trieNode[V])
		}
		next := nd.children[key[i]]
		if next == nil {
			next = &trieNode[V]{}
			nd.children[key[i]] = next
		}
		nd = next
	}
	if !nd.has {
		t.n++ // перезапись не меняет числа ключей
	}
	nd.has, nd.val = true, v
}

func (t *Trie[V]) find(key string) *trieNode[V] {
	nd := &t.root
	for i := 0; i < len(key) && nd != nil; i++ {
		nd = nd.children[key[i]]
	}
	return nd
}

// Get возвращает значение ключа и true, если ключ есть.
func (t *Trie[V]) Get(key string) (V, bool) {
	if nd := t.find(key); nd != nil && nd.has {
		return nd.val, true
	}
	var zero V
	return zero, false
}

// Len возвращает число различных ключей.
func (t *Trie[V]) Len() int { return t.n }

// WithPrefix обходит все ключи, начинающиеся с prefix (включая сам
// prefix, если он ключ), в порядке возрастания — как sort.Strings.
func (t *Trie[V]) WithPrefix(prefix string) iter.Seq2[string, V] {
	return func(yield func(string, V) bool) {
		nd := t.find(prefix)
		if nd == nil {
			return
		}
		walk(nd, []byte(prefix), yield)
	}
}

// walk — обход в глубину: сначала сам узел, потом дети по возрастанию
// байта. Возвращает false, если потребитель сделал break.
func walk[V any](nd *trieNode[V], path []byte, yield func(string, V) bool) bool {
	if nd.has && !yield(string(path), nd.val) {
		return false
	}
	for _, b := range slices.Sorted(maps.Keys(nd.children)) {
		if !walk(nd.children[b], append(path, b), yield) {
			return false
		}
	}
	return true
}
