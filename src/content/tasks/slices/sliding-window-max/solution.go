package main

// WindowMax возвращает максимумы всех окон длины k: out[i] = max(xs[i:i+k]).
// len(out) == len(xs)-k+1; если len(xs) < k — пустой результат. k <= 0 — паника.
// Время — O(n) независимо от k: у каждого элемента не больше одного
// добавления и одного удаления из дека.
func WindowMax(xs []int, k int) []int {
	if k <= 0 {
		panic("WindowMax: k должно быть больше нуля")
	}
	if len(xs) < k {
		return []int{}
	}
	out := make([]int, 0, len(xs)-k+1)
	// dq — индексы элементов окна; значения по ним строго убывают,
	// поэтому максимум окна всегда в голове.
	dq := make([]int, 0, k)
	for i, x := range xs {
		// Меньшие или равные x из хвоста уже никогда не станут максимумом.
		for len(dq) > 0 && xs[dq[len(dq)-1]] <= x {
			dq = dq[:len(dq)-1]
		}
		dq = append(dq, i)
		// Голова выпала из окна [i-k+1, i].
		if dq[0] <= i-k {
			dq = dq[1:]
		}
		if i >= k-1 {
			out = append(out, xs[dq[0]])
		}
	}
	return out
}
