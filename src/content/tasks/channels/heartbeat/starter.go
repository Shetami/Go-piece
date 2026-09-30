package main

import (
	"context"
	"time"
)

// Work запускает воркер: он берёт задачи из jobs, считает f и отдаёт
// результаты в results в порядке задач. Параллельно раз в pulse воркер
// сигналит в heartbeat «я жив» — и когда ждёт задач, и когда ждёт, пока
// заберут результат. Сигнал не блокирует воркер: если heartbeat никто не
// читает, сигнал теряется.
//
// Когда jobs закрыт и все результаты отданы или когда отменён ctx,
// воркер закрывает оба канала и завершается.
func Work(ctx context.Context, pulse time.Duration, jobs <-chan int, f func(int) int) (heartbeat <-chan struct{}, results <-chan int) {
	// ваш код
	return nil, nil
}
