package main

import (
	"cmp"
	"iter"
)

type bstNode[K cmp.Ordered, V any] struct {
	key         K
	val         V
	left, right *bstNode[K, V]
}

// Tree — двоичное дерево поиска (без балансировки). Нулевое значение
// готово к работе.
type Tree[K cmp.Ordered, V any] struct {
	root *bstNode[K, V]
	n    int
}

// Put записывает значение; существующий ключ перезаписывается.
func (t *Tree[K, V]) Put(key K, val V) {
	p := &t.root // указатель на ссылку: не нужно помнить родителя
	for *p != nil {
		switch c := cmp.Compare(key, (*p).key); {
		case c < 0:
			p = &(*p).left
		case c > 0:
			p = &(*p).right
		default:
			(*p).val = val
			return
		}
	}
	*p = &bstNode[K, V]{key: key, val: val}
	t.n++
}

// Get возвращает значение ключа и true, если ключ есть.
func (t *Tree[K, V]) Get(key K) (V, bool) {
	for nd := t.root; nd != nil; {
		switch c := cmp.Compare(key, nd.key); {
		case c < 0:
			nd = nd.left
		case c > 0:
			nd = nd.right
		default:
			return nd.val, true
		}
	}
	var zero V
	return zero, false
}

// Delete удаляет ключ и возвращает true, если он был.
func (t *Tree[K, V]) Delete(key K) bool {
	p := &t.root
	for *p != nil {
		nd := *p
		switch c := cmp.Compare(key, nd.key); {
		case c < 0:
			p = &nd.left
		case c > 0:
			p = &nd.right
		default:
			switch {
			case nd.left == nil:
				*p = nd.right
			case nd.right == nil:
				*p = nd.left
			default:
				// Два ребёнка: на место узла встаёт наименьший из правого
				// поддерева, а сам он вырезается оттуда.
				s := &nd.right
				for (*s).left != nil {
					s = &(*s).left
				}
				succ := *s
				*s = succ.right
				succ.left, succ.right = nd.left, nd.right
				*p = succ
			}
			t.n--
			return true
		}
	}
	return false
}

// Len возвращает число ключей.
func (t *Tree[K, V]) Len() int { return t.n }

// All обходит все пары по возрастанию ключа.
func (t *Tree[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		t.root.walk(nil, nil, yield)
	}
}

// Range обходит пары с lo <= key < hi по возрастанию ключа и не
// заходит в поддеревья, где подходящих ключей быть не может.
func (t *Tree[K, V]) Range(lo, hi K) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		t.root.walk(&lo, &hi, yield)
	}
}

// walk — симметричный обход с отсечением; false значит «потребитель
// сделал break, остановиться на всех уровнях». lo и hi — nil, если
// границы нет.
func (nd *bstNode[K, V]) walk(lo, hi *K, yield func(K, V) bool) bool {
	if nd == nil {
		return true
	}
	// Слева ключи меньше nd.key: идём туда, только если nd.key > lo.
	goLeft := lo == nil || cmp.Compare(*lo, nd.key) < 0
	// Справа ключи больше nd.key: идём туда, только если nd.key < hi.
	goRight := hi == nil || cmp.Compare(nd.key, *hi) < 0
	inRange := (lo == nil || cmp.Compare(*lo, nd.key) <= 0) && goRight

	if goLeft && !nd.left.walk(lo, hi, yield) {
		return false
	}
	if inRange && !yield(nd.key, nd.val) {
		return false
	}
	if goRight {
		return nd.right.walk(lo, hi, yield)
	}
	return true
}
