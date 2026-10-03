package main

type Item struct {
	Weight int
	Value  int
}

// MaxValue возвращает наибольшую суммарную ценность предметов, которые
// влезают в рюкзак вместимостью maximumWeight. Каждый предмет — один раз.
func MaxValue(maximumWeight int, items []Item) int {
	// best[w] — лучшая ценность при вместимости w среди уже рассмотренных предметов.
	best := make([]int, maximumWeight+1)
	for _, it := range items {
		// Сверху вниз: best[w-it.Weight] ещё не учитывает этот предмет,
		// поэтому взять его второй раз не получится.
		for w := maximumWeight; w >= it.Weight; w-- {
			best[w] = max(best[w], best[w-it.Weight]+it.Value)
		}
	}
	return best[maximumWeight]
}
