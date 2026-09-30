package main

import (
	"context"
	"errors"
	"sync"
)

var ErrClosed = errors.New("queue closed")

// Queue — очередь задач с фиксированным числом воркеров и буфером.
type Queue struct {
	tasks chan func()
	quit  chan struct{} // закрыт — Shutdown начат
	done  chan struct{} // закрыт — все воркеры завершились

	mu       sync.RWMutex // RLock — отправка в tasks, Lock — закрытие tasks
	closed   bool
	shutOnce sync.Once
	wg       sync.WaitGroup
}

// NewQueue запускает workers воркеров; в буфере помещается size задач.
func NewQueue(workers, size int) *Queue {
	q := &Queue{
		tasks: make(chan func(), size),
		quit:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	q.wg.Add(workers)
	for range workers {
		go func() {
			defer q.wg.Done()
			for task := range q.tasks { // разбирает буфер до конца даже после Shutdown
				task()
			}
		}()
	}
	return q
}

// Submit ставит задачу в очередь. Если буфер полон — ждёт места.
// Возвращает ctx.Err(), если ctx отменён раньше; ErrClosed, если Shutdown
// уже начат (в том числе пока Submit ждал места). Безопасен одновременно
// с Shutdown. Принятая задача (nil) обязательно будет выполнена.
func (q *Queue) Submit(ctx context.Context, task func()) error {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.closed {
		return ErrClosed
	}
	select {
	case q.tasks <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-q.quit: // Shutdown ждёт наш RUnlock — не держим его вечно
		return ErrClosed
	}
}

// Shutdown перестаёт принимать задачи, дожидается выполнения всех уже
// принятых (включая стоящие в буфере) и остановки воркеров.
// Если ctx истёк раньше — возвращает ctx.Err(), воркеры доделывают в фоне.
// Повторные и одновременные вызовы безопасны.
func (q *Queue) Shutdown(ctx context.Context) error {
	q.shutOnce.Do(func() {
		close(q.quit) // будим тех, кто ждёт места в буфере
		q.mu.Lock()   // ждём, пока все Submit выйдут из отправки
		q.closed = true
		close(q.tasks)
		q.mu.Unlock()
		go func() {
			q.wg.Wait()
			close(q.done)
		}()
	})
	select {
	case <-q.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
