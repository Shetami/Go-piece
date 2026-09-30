package main

import "context"

// RunPriority запускает workers воркеров, которые выполняют задачи из двух
// очередей: high и low.
//
//   - Если в high есть задача, воркер берёт её раньше любой задачи из low.
//   - Закрытие одной очереди не останавливает работу с другой; RunPriority
//     возвращает nil, когда обе очереди закрыты и все задачи выполнены.
//   - При отмене ctx воркеры не начинают новых задач; RunPriority
//     дожидается уже выполняющихся и возвращает ctx.Err().
func RunPriority(ctx context.Context, workers int, high, low <-chan func()) error {
	// ваш код
	return nil
}
