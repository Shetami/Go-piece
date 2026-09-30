package main

import "errors"

var (
	ErrNotFound  = errors.New("не найдено")
	ErrForbidden = errors.New("доступ запрещён")
	ErrConflict  = errors.New("конфликт")
)

// ValidationError — клиент прислал некорректные данные.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Msg }

// Respond превращает ошибку обработчика в HTTP-статус и тело ответа.
// Правила проверяются сверху вниз, срабатывает первое подходящее;
// ошибки могут быть обёрнуты и объединены через errors.Join.
//
//	nil                        → 200, ""
//	*ValidationError в цепочке → 400, текст этой ValidationError (не всей ошибки)
//	ErrNotFound                → 404, "не найдено"
//	ErrForbidden               → 403, "доступ запрещён"
//	ErrConflict                → 409, "конфликт"
//	context.DeadlineExceeded   → 504, "таймаут"
//	context.Canceled           → 499, "запрос отменён"
//	всё остальное              → 500, "внутренняя ошибка" (никаких подробностей)
func Respond(err error) (int, string) {
	// ваш код
	return 500, "внутренняя ошибка"
}
