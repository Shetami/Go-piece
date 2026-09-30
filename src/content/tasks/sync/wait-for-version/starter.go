package main

import "context"

// Version — номер последней применённой версии (например, позиции в
// журнале репликации). Запросы «прочитай свои записи» ждут, пока реплика
// догонит нужную версию.
type Version struct {
	// ваши поля
}

func NewVersion() *Version {
	// ваш код
	return &Version{}
}

// Set сообщает о новой версии. Версия только растёт: значение меньше
// или равное текущему игнорируется.
func (v *Version) Set(n int64) {
	// ваш код
}

// Get возвращает текущую версию.
func (v *Version) Get() int64 {
	// ваш код
	return 0
}

// WaitFor блокируется, пока версия не станет >= n, и возвращает nil.
// Если ctx отменён раньше — возвращает ctx.Err(). Если версия уже
// достигнута, возвращает nil сразу, даже при отменённом ctx.
func (v *Version) WaitFor(ctx context.Context, n int64) error {
	// ваш код
	return nil
}
