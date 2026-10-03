package main

import "slices"

// Compare сравнивает два списка и возвращает:
//
//	"equal"     — списки одинаковы;
//	"sublist"   — a целиком, подряд и по порядку содержится в b;
//	"superlist" — b содержится в a;
//	"unequal"   — ничего из этого.
func Compare(listOne, listTwo []int) string {
	switch {
	case slices.Equal(listOne, listTwo):
		return "equal"
	case contains(listTwo, listOne):
		return "sublist"
	case contains(listOne, listTwo):
		return "superlist"
	}
	return "unequal"
}

// contains сообщает, что part встречается в whole подряд идущим куском.
func contains(whole, part []int) bool {
	for i := 0; i+len(part) <= len(whole); i++ {
		if slices.Equal(whole[i:i+len(part)], part) {
			return true
		}
	}
	return false
}
