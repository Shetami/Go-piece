package main

import "sync"

// OnceValues возвращает функцию, которая при первом вызове вызывает f,
// запоминает её результат и дальше всегда возвращает его.
//   - Безопасна для горутин: f вызывается ровно один раз, одновременные
//     вызовы ждут, пока первый закончит.
//   - Ошибка запоминается так же, как значение: повторных попыток нет.
//   - Если f запаниковала, паникует каждый вызов — и первый, и все
//     последующие — с тем же значением паники; f повторно не вызывается.
func OnceValues[T any](f func() (T, error)) func() (T, error) {
	var (
		once     sync.Once
		val      T
		err      error
		panicked bool
		pval     any
	)
	return func() (T, error) {
		once.Do(func() {
			// Флаг ставим до вызова и снимаем после: если f паникует,
			// до снятия дело не дойдёт, и defer поймёт, что была паника.
			panicked = true
			defer func() {
				if panicked {
					pval = recover()
				}
			}()
			val, err = f()
			f = nil // f больше не нужна — отпускаем то, что она захватила
			panicked = false
		})
		// once.Do даёт happens-before: всё, что записано внутри, видно здесь.
		if panicked {
			panic(pval)
		}
		return val, err
	}
}
