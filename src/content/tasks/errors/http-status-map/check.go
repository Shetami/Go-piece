package main

func chkRespond(t *testing.T, name string, err error, code int, body string) {
	t.Helper()
	c, b := Respond(err)
	if c != code || b != body {
		t.Fatalf("%s: Respond = (%d, %q), ожидали (%d, %q)", name, c, b, code, body)
	}
}

func TestRespondBasic(t *testing.T) {
	chkRespond(t, "nil", nil, 200, "")
	chkRespond(t, "не найдено", fmt.Errorf("заказ 7: %w", ErrNotFound), 404, "не найдено")
	chkRespond(t, "запрещено", fmt.Errorf("a: %w", fmt.Errorf("b: %w", ErrForbidden)), 403, "доступ запрещён")
	chkRespond(t, "конфликт", fmt.Errorf("обновить: %w", ErrConflict), 409, "конфликт")
}

func TestRespondValidation(t *testing.T) {
	err := fmt.Errorf("handler: сохранить заказ: %w", &ValidationError{"email", "нет @"})
	chkRespond(t, "валидация", err, 400, "email: нет @")
}

func TestRespondContext(t *testing.T) {
	chkRespond(t, "дедлайн", fmt.Errorf("запрос в базу: %w", context.DeadlineExceeded), 504, "таймаут")
	chkRespond(t, "отмена", fmt.Errorf("запрос в базу: %w", context.Canceled), 499, "запрос отменён")
}

func TestRespondPriority(t *testing.T) {
	err := errors.Join(fmt.Errorf("x: %w", ErrNotFound), &ValidationError{"id", "не число"})
	chkRespond(t, "Join валидации и 404", err, 400, "id: не число")
	err = errors.Join(context.Canceled, fmt.Errorf("y: %w", ErrConflict))
	chkRespond(t, "Join отмены и конфликта", err, 409, "конфликт")
}

func TestRespondNoLeak(t *testing.T) {
	err := fmt.Errorf("pq: password authentication failed for user \"admin\"")
	chkRespond(t, "неизвестная ошибка", err, 500, "внутренняя ошибка")
}
