package main

import "slices"

// Account — учётная запись: имя и адреса почты.
type Account struct {
	Name   string
	Emails []string
}

// MergeAccounts склеивает учётные записи одного человека.
//
//   - Две записи принадлежат одному человеку, если у них есть общий email,
//     в том числе через цепочку: A и B делят один адрес, B и C — другой.
//   - Одинаковое имя само по себе ничего не значит: тёзки — разные люди.
//   - В результате у группы имя первой (по входу) записи группы, адреса —
//     без повторов и по алфавиту. Группы идут в порядке своей первой записи.
//   - Запись без адресов — отдельная группа.
//   - Входной слайс и его Emails не меняются.
func MergeAccounts(accs []Account) []Account {
	// Система непересекающихся множеств над индексами записей.
	parent := make([]int, len(accs))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x]) // сжатие пути
		}
		return parent[x]
	}
	union := func(a, b int) {
		ra, rb := find(a), find(b)
		if ra == rb {
			return
		}
		// Корень — меньший индекс: он же первая запись группы.
		if rb < ra {
			ra, rb = rb, ra
		}
		parent[rb] = ra
	}

	owner := make(map[string]int) // email → первая запись, где он встретился
	for i, a := range accs {
		for _, e := range a.Emails {
			if j, ok := owner[e]; ok {
				union(i, j)
			} else {
				owner[e] = i
			}
		}
	}

	pos := make(map[int]int) // корень → индекс группы в ответе
	var res []Account
	for i, a := range accs {
		r := find(i)
		p, ok := pos[r]
		if !ok {
			p = len(res)
			pos[r] = p
			res = append(res, Account{Name: accs[r].Name})
		}
		// append к новому слайсу — входные Emails не трогаем.
		res[p].Emails = append(res[p].Emails, a.Emails...)
	}
	for i := range res {
		slices.Sort(res[i].Emails)
		res[i].Emails = slices.Compact(res[i].Emails)
	}
	return res
}
