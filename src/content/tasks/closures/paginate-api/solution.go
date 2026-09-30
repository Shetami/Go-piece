package main

import (
	"context"
	"errors"
	"fmt"
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
	return func(yield func(T, error) bool) {
		var zero T
		// Состояние обхода живёт внутри функции-итератора, а не в All:
		// иначе второй range продолжил бы с того места, где кончил первый.
		cursor := ""
		seen := make(map[string]bool)
		for {
			if err := ctx.Err(); err != nil {
				yield(zero, err)
				return
			}
			if seen[cursor] {
				yield(zero, fmt.Errorf("%w: %q", ErrCursorLoop, cursor))
				return
			}
			seen[cursor] = true

			page, err := fetch(ctx, cursor)
			if err != nil {
				yield(zero, fmt.Errorf("страница %q: %w", cursor, err))
				return
			}
			for _, it := range page.Items {
				if !yield(it, nil) {
					return // break у потребителя: дальше не запрашиваем
				}
			}
			if page.Next == "" {
				return
			}
			cursor = page.Next // пустые Items сюда тоже доходят — это не конец
		}
	}
}
