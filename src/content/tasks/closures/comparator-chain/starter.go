package main

import "cmp"

// By склеивает компараторы: сравнивает по первому, при равенстве — по
// второму и так далее. Без аргументов все элементы равны (возвращает 0).
// Компаратор — как в slices.SortFunc: <0, 0 или >0.
func By[T any](cmps ...func(a, b T) int) func(a, b T) int {
	// ваш код
	return func(a, b T) int { return 0 }
}

// Key возвращает компаратор по ключу: key(a) против key(b).
func Key[T any, K cmp.Ordered](key func(T) K) func(a, b T) int {
	// ваш код
	return func(a, b T) int { return 0 }
}

// Desc разворачивает порядок компаратора.
func Desc[T any](c func(a, b T) int) func(a, b T) int {
	// ваш код
	return c
}
