package main

// CanChain сообщает, можно ли выложить все кости домино в замкнутую цепочку:
// соседние половинки совпадают, и свободные концы первой и последней кости тоже.
// Кость можно переворачивать.
func CanChain(dominoes [][2]int) bool {
	if len(dominoes) == 0 {
		return true
	}
	// Кость — ребро между числами на её половинках, цепочка — эйлеров цикл.
	// Он есть, когда степень каждой вершины чётна и все рёбра в одной компоненте.
	degree := map[int]int{}
	parent := map[int]int{}
	var find func(int) int
	find = func(x int) int {
		if _, ok := parent[x]; !ok {
			parent[x] = x
		}
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	for _, d := range dominoes {
		degree[d[0]]++
		degree[d[1]]++
		parent[find(d[0])] = find(d[1])
	}
	root := find(dominoes[0][0])
	for v, deg := range degree {
		if deg%2 != 0 || find(v) != root {
			return false
		}
	}
	return true
}
