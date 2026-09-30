package main

// Weight — типы весов рёбер, в том числе свои (type Millis int64).
type Weight interface {
	~int | ~int64 | ~float64
}

// Graph — ориентированный взвешенный граф. Нулевое значение готово к работе.
type Graph[N comparable, W Weight] struct {
	// ваши поля
}

// AddEdge добавляет ребро from → to с весом w >= 0. Параллельные рёбра
// допустимы.
func (g *Graph[N, W]) AddEdge(from, to N, w W) {
	// ваш код
}

// ShortestPath возвращает путь минимального суммарного веса от from до
// to (обе вершины включительно), его вес и true. Если пути нет — nil,
// 0 и false. Путь от вершины к самой себе — [from], 0, true.
func (g *Graph[N, W]) ShortestPath(from, to N) ([]N, W, bool) {
	// ваш код
	return nil, 0, false
}
