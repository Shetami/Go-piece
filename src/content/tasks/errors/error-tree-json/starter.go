package main

// Node — ошибка в виде дерева для JSON-ответа или структурного лога.
type Node struct {
	Msg    string `json:"msg"`
	Code   string `json:"code,omitempty"`
	Causes []Node `json:"causes,omitempty"`
}

// Coder — ошибки с машинным кодом.
type Coder interface{ Code() string }

// MaxDepth — сколько уровней дерева выводить; корень — уровень 1.
const MaxDepth = 8

// Encode раскладывает err в дерево:
//   - nil → nil;
//   - Msg — Error() текущего звена;
//   - Code — Code() текущего звена, если ОНО САМО реализует Coder (коды
//     глубже по цепочке к родителям не поднимаются);
//   - Causes — дети звена: результат Unwrap() error (если не nil) либо по
//     одному ребёнку на каждый не-nil элемент Unwrap() []error, в том же
//     порядке;
//   - узел на уровне MaxDepth выводится без Causes: ошибка, которая
//     разворачивается сама в себя, не должна уронить сервис.
func Encode(err error) *Node {
	// ваш код
	return nil
}
