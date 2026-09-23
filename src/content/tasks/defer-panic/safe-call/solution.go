package main

import "fmt"

// SafeCall вызывает f и превращает панику в ошибку.
// Если f паникует ошибкой — она должна находиться через errors.Is/As.
// Если чем-то другим — в тексте ошибки должно быть это значение.
// Без паники — nil.
func SafeCall(f func()) (err error) {
	defer func() {
		switch r := recover().(type) {
		case nil:
		case error:
			// %w сохраняет цепочку: снаружи работают errors.Is и errors.As.
			err = fmt.Errorf("паника: %w", r)
		default:
			err = fmt.Errorf("паника: %v", r)
		}
	}()
	f()
	return nil
}
