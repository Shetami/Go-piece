package main

// Conn — соединение, которое закрывают из нескольких мест: из defer
// обработчика, из обработчика сигнала, из watchdog'а.
type Conn struct {
	closeFn func() error
	// ваши поля
}

func NewConn(closeFn func() error) *Conn {
	// ваш код
	return &Conn{closeFn: closeFn}
}

// Close закрывает соединение.
//   - closeFn вызывается ровно один раз, даже при одновременных Close.
//   - Все вызовы Close — одновременные и последующие — возвращают то же,
//     что вернул closeFn. Одновременные ждут, пока closeFn завершится.
//   - Если closeFn запаниковала — Close не паникует, а возвращает ошибку
//     с текстом паники, и все остальные вызовы возвращают ту же ошибку.
func (c *Conn) Close() error {
	// ваш код
	return c.closeFn()
}

// Done возвращает канал, который закрывается, когда closeFn завершилась
// (успешно, с ошибкой или паникой).
func (c *Conn) Done() <-chan struct{} {
	// ваш код
	return nil
}
