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
	if err == nil {
		return nil
	}
	n := encode(err, 1)
	return &n
}

func encode(err error, depth int) Node {
	n := Node{Msg: err.Error()}
	// Утверждение типа, а не errors.As: As нашёл бы код где-то в глубине
	// и приписал его каждому предку.
	if c, ok := err.(Coder); ok {
		n.Code = c.Code()
	}
	if depth >= MaxDepth {
		return n
	}
	// Смотрим оба вида Unwrap: errors.Unwrap видит только одиночный и
	// для Join и fmt.Errorf с несколькими %w вернёт nil.
	switch u := err.(type) {
	case interface{ Unwrap() error }:
		if c := u.Unwrap(); c != nil {
			n.Causes = []Node{encode(c, depth+1)}
		}
	case interface{ Unwrap() []error }:
		for _, c := range u.Unwrap() {
			if c != nil {
				n.Causes = append(n.Causes, encode(c, depth+1))
			}
		}
	}
	return n
}
