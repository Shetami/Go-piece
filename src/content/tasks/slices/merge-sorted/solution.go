package main

// MergeSorted сливает два отсортированных по возрастанию слайса
// в один отсортированный. Входные слайсы не меняются.
func MergeSorted(a, b []int) []int {
	out := make([]int, 0, len(a)+len(b))
	i, j := 0, 0
	// Два указателя: на каждом шаге берём меньший из двух голов.
	for i < len(a) && j < len(b) {
		if a[i] <= b[j] {
			out = append(out, a[i])
			i++
		} else {
			out = append(out, b[j])
			j++
		}
	}
	// Один из слайсов кончился — хвост другого уже отсортирован.
	out = append(out, a[i:]...)
	out = append(out, b[j:]...)
	return out
}
