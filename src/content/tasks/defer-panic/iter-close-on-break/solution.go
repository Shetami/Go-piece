package main

import "iter"

// Source — открытый источник строк (файл, курсор базы, поток из сети).
type Source interface {
	// Next возвращает следующую строку; ok=false — строки кончились.
	Next() (line string, ok bool, err error)
	Close() error
}

// Lines возвращает итератор по строкам источника.
//   - open вызывается, когда начинается range, а не при вызове Lines;
//     каждый новый range открывает источник заново.
//   - Источник закрывается ровно один раз на каждый range: когда строки
//     кончились, когда потребитель вышел из цикла (break, return), когда
//     тело цикла запаниковало (паника летит дальше) и после ошибки.
//   - Ошибка open или Next отдаётся одной парой ("", err), после неё
//     итерация заканчивается.
//   - Ошибка Close после того, как строки нормально кончились, отдаётся
//     последней парой ("", err). После break её отдавать некому.
func Lines(open func() (Source, error)) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		src, err := open()
		if err != nil {
			yield("", err)
			return
		}

		// Страховка для break и паники в теле цикла: паника из yield
		// проходит через этот кадр, и отложенный Close выполнится.
		closed := false
		defer func() {
			if !closed {
				_ = src.Close() // ошибку отдать уже некому
			}
		}()

		for {
			line, ok, err := src.Next()
			if err != nil {
				closed = true
				_ = src.Close()
				yield("", err)
				return
			}
			if !ok {
				closed = true
				if cerr := src.Close(); cerr != nil {
					yield("", cerr) // потребитель ещё слушает — можно
				}
				return
			}
			if !yield(line, nil) {
				return // break: больше yield вызывать нельзя
			}
		}
	}
}
