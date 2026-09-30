package main

import (
	"cmp"
	"slices"
	"strings"
)

// Suggest подбирает подсказки «возможно, вы имели в виду» для опечатки
// в команде CLI или поисковом запросе.
//
//   - расстояние — Левенштейна (вставка, удаление, замена по 1), считается
//     в символах (рунах) и без учёта регистра (strings.ToLower);
//   - подходят кандидаты с расстоянием <= maxDist;
//   - кандидаты, совпадающие без учёта регистра, — один вариант: первое
//     написание из candidates;
//   - порядок: по возрастанию расстояния, при равенстве — по
//     strings.Compare исходного написания;
//   - возвращается не больше limit вариантов; ничего не подошло — nil.
func Suggest(input string, candidates []string, maxDist, limit int) []string {
	type hit struct {
		s    string
		dist int
	}
	in := []rune(strings.ToLower(input))
	seen := map[string]bool{}
	var hits []hit
	for _, c := range candidates {
		lc := strings.ToLower(c)
		if seen[lc] {
			continue
		}
		seen[lc] = true
		if d := levenshtein(in, []rune(lc)); d <= maxDist {
			hits = append(hits, hit{c, d})
		}
	}
	slices.SortFunc(hits, func(a, b hit) int {
		return cmp.Or(cmp.Compare(a.dist, b.dist), strings.Compare(a.s, b.s))
	})
	var out []string
	for _, h := range hits[:min(limit, len(hits))] {
		out = append(out, h.s)
	}
	return out
}

// levenshtein — классическая динамика на двух строках таблицы: O(n·m)
// по времени и O(m) по памяти.
func levenshtein(a, b []rune) int {
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}
