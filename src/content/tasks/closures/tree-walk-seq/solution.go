package main

import "iter"

// Node — узел дерева (каталог, категория, элемент меню).
type Node struct {
	Name     string
	Children []*Node
}

// Walk обходит дерево в прямом порядке (узел, потом его дети по порядку)
// и отдаёт путь от корня — имена, включая сам узел, — и узел.
//   - skip != nil и skip(n) == true: ни узел, ни его поддерево не отдаются
//     (skip для потомков пропущенного узла не вызывается).
//   - root == nil — пустая последовательность.
//   - Каждый отданный путь — отдельный слайс: потребитель может его хранить.
//   - break на любой глубине останавливает обход целиком.
func Walk(root *Node, skip func(*Node) bool) iter.Seq2[[]string, *Node] {
	return func(yield func([]string, *Node) bool) {
		// Рекурсивное замыкание: объявить, потом присвоить.
		var walk func(n *Node, parent []string) bool
		walk = func(n *Node, parent []string) bool {
			if skip != nil && skip(n) {
				return true
			}
			// [:len:len] — append всегда в новый массив: братья не
			// перетирают путь друг друга и уже отданные пути.
			path := append(parent[:len(parent):len(parent)], n.Name)
			if !yield(path, n) {
				return false
			}
			for _, c := range n.Children {
				// false изнутри рекурсии должен подняться до самого верха.
				if !walk(c, path) {
					return false
				}
			}
			return true
		}
		if root != nil {
			walk(root, nil)
		}
	}
}
