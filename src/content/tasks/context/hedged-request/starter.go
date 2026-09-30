package main

import (
	"context"
	"errors"
	"time"
)

var ErrNoReplicas = errors.New("hedged: нет реплик")

// Replica — запрос к одной реплике. Обязана слушать ctx.
type Replica func(ctx context.Context) (string, error)

// Hedged — запрос «с подстраховкой». Сначала запускается replicas[0].
// Если за delay ответа нет — запускается следующая реплика, ещё через
// delay — следующая, и так далее; предыдущие при этом продолжают работать.
// Если реплика вернула ошибку — следующая запускается сразу, не дожидаясь
// delay. Первый успешный ответ возвращается немедленно, а контекст всех
// остальных запущенных реплик отменяется.
// Все реплики упали — ошибка errors.Join всех их ошибок (каждая с номером
// реплики). Отменили ctx — context.Cause(ctx) сразу.
// Горутины реплик не должны оставаться висеть после выхода.
func Hedged(ctx context.Context, delay time.Duration, replicas []Replica) (string, error) {
	// ваш код
	return "", ErrNoReplicas
}
