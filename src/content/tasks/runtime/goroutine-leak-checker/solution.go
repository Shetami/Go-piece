package main

import (
	"fmt"
	"runtime"
	"time"
)

// Snapshot запоминает, сколько горутин живо сейчас, и возвращает функцию
// проверки — как goleak или самописный хелпер в тестах.
//
// check(timeout) ждёт, пока число горутин не вернётся к запомненному
// (или не станет меньше): горутины, которые завершаются чуть позже, не
// считаются утечкой. Если за timeout этого не произошло, check возвращает
// ошибку, текст которой начинается с "goroutine leak" и содержит полный
// дамп стеков всех горутин — без обрезки, сколько бы их ни было.
// Сама check горутин не оставляет и не ждёт заметно дольше timeout.
func Snapshot() (check func(timeout time.Duration) error) {
	base := runtime.NumGoroutine()
	return func(timeout time.Duration) error {
		deadline := time.Now().Add(timeout)
		// Опрос с паузой: горутине, которая вот-вот завершится,
		// нужно дать время. Без таймеров и без новых горутин.
		for {
			n := runtime.NumGoroutine()
			if n <= base {
				return nil
			}
			if !time.Now().Before(deadline) {
				return fmt.Errorf("goroutine leak: было %d, стало %d\n%s", base, n, allStacks())
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
}

// allStacks возвращает дамп всех горутин. runtime.Stack молча обрезает
// вывод по размеру буфера, поэтому растим буфер, пока дамп не влезет.
func allStacks() []byte {
	buf := make([]byte, 64<<10)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}
