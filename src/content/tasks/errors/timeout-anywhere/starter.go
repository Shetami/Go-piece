package main

// Kind — что случилось с запросом, с точки зрения метрик и алертов.
type Kind int

const (
	KindOK       Kind = iota // ошибки нет
	KindTimeout              // мы не уложились во время
	KindCanceled             // клиент ушёл сам
	KindFailure              // всё остальное
)

// IsTimeout сообщает, есть ли ГДЕ-НИБУДЬ в дереве err звено с методом
// Timeout() bool, вернувшим true. Обходится всё дерево — и цепочки
// Unwrap() error, и все ветки Unwrap() []error. Звено с Timeout() == false
// поиск не прекращает: ответ может быть глубже.
func IsTimeout(err error) bool {
	// ваш код
	return false
}

// Classify: nil → KindOK; IsTimeout → KindTimeout; иначе
// errors.Is(err, context.Canceled) → KindCanceled; иначе KindFailure.
// Таймаут важнее отмены: если в дереве есть и то и другое — KindTimeout.
func Classify(err error) Kind {
	// ваш код
	return KindFailure
}
