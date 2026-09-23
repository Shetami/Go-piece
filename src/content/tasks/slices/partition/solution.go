package main

// Partition делит элементы на подходящие под pred и остальные.
// Порядок внутри обеих частей сохраняется, исходный слайс не меняется.
func Partition[T any](xs []T, pred func(T) bool) (yes, no []T) {
	// Два новых слайса, а не xs[:0]: переиспользование исходного массива
	// испортило бы вход вызывающему.
	for _, x := range xs {
		if pred(x) {
			yes = append(yes, x)
		} else {
			no = append(no, x)
		}
	}
	return yes, no
}
