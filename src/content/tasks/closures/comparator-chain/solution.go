package main

import "cmp"

// By склеивает компараторы: сравнивает по первому, при равенстве — по
// второму и так далее. Без аргументов все элементы равны (возвращает 0).
// Компаратор — как в slices.SortFunc: <0, 0 или >0.
func By[T any](cmps ...func(a, b T) int) func(a, b T) int {
	// Копия: вызывающий может поменять свой слайс после сборки.
	cmps = append([]func(a, b T) int(nil), cmps...)
	return func(a, b T) int {
		for _, c := range cmps {
			if r := c(a, b); r != 0 {
				return r
			}
		}
		return 0
	}
}

// Key возвращает компаратор по ключу: key(a) против key(b).
func Key[T any, K cmp.Ordered](key func(T) K) func(a, b T) int {
	return func(a, b T) int {
		return cmp.Compare(key(a), key(b))
	}
}

// Desc разворачивает порядок компаратора.
func Desc[T any](c func(a, b T) int) func(a, b T) int {
	// Аргументы местами, а не -c(a, b): минус от math.MinInt — снова MinInt.
	return func(a, b T) int {
		return c(b, a)
	}
}
