package main

import (
	"container/heap"
	"slices"
)

// Weight — типы весов рёбер, в том числе свои (type Millis int64).
type Weight interface {
	~int | ~int64 | ~float64
}

type edge[N comparable, W Weight] struct {
	to N
	w  W
}

// Graph — ориентированный взвешенный граф. Нулевое значение готово к работе.
type Graph[N comparable, W Weight] struct {
	adj map[N][]edge[N, W]
}

// AddEdge добавляет ребро from → to с весом w >= 0. Параллельные рёбра
// допустимы.
func (g *Graph[N, W]) AddEdge(from, to N, w W) {
	if g.adj == nil {
		g.adj = make(map[N][]edge[N, W])
	}
	g.adj[from] = append(g.adj[from], edge[N, W]{to, w})
}

// item и pq — очередь для container/heap. Обобщённый тип тоже может
// реализовать heap.Interface.
type item[N comparable, W Weight] struct {
	node N
	dist W
}

type pq[N comparable, W Weight] []item[N, W]

func (q pq[N, W]) Len() int           { return len(q) }
func (q pq[N, W]) Less(i, j int) bool { return q[i].dist < q[j].dist }
func (q pq[N, W]) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *pq[N, W]) Push(x any)        { *q = append(*q, x.(item[N, W])) }
func (q *pq[N, W]) Pop() any {
	old := *q
	it := old[len(old)-1]
	*q = old[:len(old)-1]
	return it
}

// ShortestPath возвращает путь минимального суммарного веса от from до
// to (обе вершины включительно), его вес и true. Если пути нет — nil,
// 0 и false. Путь от вершины к самой себе — [from], 0, true.
func (g *Graph[N, W]) ShortestPath(from, to N) ([]N, W, bool) {
	dist := map[N]W{from: 0}
	prev := map[N]N{}
	done := map[N]bool{}
	q := &pq[N, W]{{from, 0}}
	for q.Len() > 0 {
		cur := heap.Pop(q).(item[N, W])
		if done[cur.node] {
			continue // устаревшая запись: вершину уже достали с меньшим весом
		}
		done[cur.node] = true
		if cur.node == to {
			break // вес to больше не уменьшится
		}
		for _, e := range g.adj[cur.node] {
			nd := cur.dist + e.w
			if d, seen := dist[e.to]; !seen || nd < d {
				dist[e.to] = nd
				prev[e.to] = cur.node
				// decrease-key нет — просто кладём ещё одну запись.
				heap.Push(q, item[N, W]{e.to, nd})
			}
		}
	}
	if !done[to] {
		return nil, 0, false
	}
	path := []N{to}
	for n := to; n != from; {
		n = prev[n]
		path = append(path, n)
	}
	slices.Reverse(path)
	return path, dist[to], true
}
