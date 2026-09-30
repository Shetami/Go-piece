package main

import (
	"sync"
	"sync/atomic"
)

// Node — узел дерева с произвольным числом детей.
type Node struct {
	Val      int
	Children []*Node
}

// Sum возвращает сумму weight(n) по всем узлам дерева с корнем root
// (nil — пустое дерево, 0).
//
// weight может быть медленной (в тестах — спит), поэтому поддеревья
// обходятся параллельно, но одновременно выполняется не больше limit
// вызовов weight (limit >= 1, вызывающая горутина считается).
// Дерево может быть очень глубоким (цепочка из тысяч узлов) и очень
// широким (тысячи детей у одного узла).
func Sum(root *Node, limit int, weight func(*Node) int) int {
	if root == nil {
		return 0
	}
	var total atomic.Int64
	var wg sync.WaitGroup
	// Слоты для дополнительных горутин; вызывающая уже работает.
	sem := make(chan struct{}, limit-1)

	var walk func(n *Node)
	walk = func(n *Node) {
		total.Add(int64(weight(n)))
		for _, c := range n.Children {
			select {
			case sem <- struct{}{}:
				// Есть свободный слот — поддерево в новой горутине.
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer func() { <-sem }()
					walk(c)
				}()
			default:
				// Слотов нет — обходим сами. Ждать слот тут нельзя:
				// все владельцы слотов тоже могут ждать — дедлок.
				walk(c)
			}
		}
	}
	walk(root)
	wg.Wait()
	return int(total.Load())
}
