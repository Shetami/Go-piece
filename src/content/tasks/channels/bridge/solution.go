package main

import "context"

// Bridge превращает поток каналов в один канал значений: вычитывает первый
// канал до закрытия, потом второй и так далее, сохраняя порядок.
// Выход закрывается, когда закрыт streams и вычитан последний канал,
// или когда отменён ctx. После отмены горутины Bridge не остаются висеть.
func Bridge[T any](ctx context.Context, streams <-chan (<-chan T)) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for {
			// Каждое ожидание — и следующего канала, и значения, и отправки —
			// идёт в select с ctx.Done(): иначе отмена застанет горутину врасплох.
			var stream <-chan T
			select {
			case s, ok := <-streams:
				if !ok {
					return
				}
				stream = s
			case <-ctx.Done():
				return
			}
			for {
				var v T
				select {
				case x, ok := <-stream:
					if !ok {
						stream = nil
					}
					v = x
				case <-ctx.Done():
					return
				}
				if stream == nil {
					break
				}
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
