package main

// Node — вершина ориентированного графа. Имена не уникальны.
type Node struct {
	Name  string
	Edges []*Node
}

// Clone возвращает глубокую копию графа, достижимого из root.
//
//   - Ни одна вершина копии не совпадает с вершиной оригинала.
//   - Структура сохраняется точно: если в оригинале две ссылки ведут на одну
//     вершину, в копии они тоже ведут на одну; циклы и петли остаются
//     циклами и петлями; порядок и повторы в Edges сохраняются.
//   - Разные вершины с одинаковым Name остаются разными.
//   - Clone(nil) == nil. Оригинал не меняется.
func Clone(root *Node) *Node {
	if root == nil {
		return nil
	}
	// Ключ — указатель, а не имя: он и есть идентичность вершины.
	copies := map[*Node]*Node{root: {Name: root.Name}}
	stack := []*Node{root} // явный стек: длинная цепочка не переполнит рекурсию
	for len(stack) > 0 {
		orig := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		cp := copies[orig]
		cp.Edges = make([]*Node, len(orig.Edges))
		for i, next := range orig.Edges {
			if next == nil {
				continue
			}
			c, ok := copies[next]
			if !ok {
				// Копию регистрируем ДО обхода её рёбер — так цикл
				// вернётся к уже известной вершине, а не создаст новую.
				c = &Node{Name: next.Name}
				copies[next] = c
				stack = append(stack, next)
			}
			cp.Edges[i] = c
		}
	}
	return copies[root]
}
