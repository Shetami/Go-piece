package main

import "context"

// Bridge превращает поток каналов в один канал значений: вычитывает первый
// канал до закрытия, потом второй и так далее, сохраняя порядок.
// Выход закрывается, когда закрыт streams и вычитан последний канал,
// или когда отменён ctx. После отмены горутины Bridge не остаются висеть.
func Bridge[T any](ctx context.Context, streams <-chan (<-chan T)) <-chan T {
	// ваш код
	out := make(chan T)
	close(out)
	return out
}
