package main

import (
	"context"
	"errors"
	"iter"
)

// ErrCursorLoop — сервер вернул курсор, по которому мы уже ходили.
var ErrCursorLoop = errors.New("курсор зациклился")

// Page — одна страница ответа. Next == "" — страниц больше нет.
type Page[T any] struct {
	Items []T
	Next  string
}

// Fetch запрашивает страницу по курсору; пустой курсор — первая страница.
type Fetch[T any] func(ctx context.Context, cursor string) (Page[T], error)

// All лениво обходит элементы всех страниц по порядку.
//
//   - следующая страница запрашивается, только когда элементы текущей
//     кончились и потребитель хочет ещё; после break запросов больше нет;
//   - пустая страница с непустым Next — не конец, идём дальше;
//   - ошибка Fetch приходит один раз парой (нулевое значение, ошибка),
//     обёрнутой через %w, с курсором в тексте; после неё обход закончен;
//   - перед запросом, если ctx отменён, — пара (нулевое, ctx.Err());
//   - повторный курсор — пара (нулевое, ошибка с ErrCursorLoop);
//   - каждый обход результата начинается с первой страницы.
func All[T any](ctx context.Context, fetch Fetch[T]) iter.Seq2[T, error] {
	// ваш код
	return func(yield func(T, error) bool) {}
}
