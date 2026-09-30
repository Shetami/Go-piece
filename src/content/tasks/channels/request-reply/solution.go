package main

import (
	"context"
	"errors"
	"maps"
	"sync"
)

var ErrClosed = errors.New("store closed")

// Store — хранилище счётчиков, которым владеет одна горутина. Все
// операции уходят к ней запросами по каналу, ответ возвращается по каналу
// ответа из запроса. Мьютексов для данных нет — ими владеет горутина.
type Store struct {
	reqs chan func(map[string]int) // запрос — функция, выполняемая владельцем
	quit chan struct{}             // закрывается в Close
	done chan struct{}             // закрывается, когда владелец вышел
	once sync.Once
}

// NewStore создаёт хранилище и запускает горутину-владельца.
func NewStore() *Store {
	s := &Store{
		reqs: make(chan func(map[string]int)),
		quit: make(chan struct{}),
		done: make(chan struct{}),
	}
	go func() {
		defer close(s.done)
		data := make(map[string]int) // живёт только в этой горутине
		for {
			select {
			case req := <-s.reqs:
				req(data)
			case <-s.quit:
				return
			}
		}
	}()
	return s
}

// do отправляет запрос владельцу. Запрос сам пишет ответ в свой
// канал с буфером на один, поэтому владелец никогда не ждёт вызывающего.
func (s *Store) do(ctx context.Context, req func(map[string]int)) error {
	select {
	case s.reqs <- req:
		return nil
	case <-s.quit:
		return ErrClosed
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Update применяет fn к счётчику key (отсутствующий считается нулём)
// в горутине-владельце и возвращает новое значение. Если ctx отменён,
// пока запрос ждёт отправки или ответа, — ctx.Err() (изменение при этом
// может успеть примениться). Медленный или ушедший вызывающий не должен
// останавливать обработку чужих запросов. После Close — ErrClosed.
func (s *Store) Update(ctx context.Context, key string, fn func(old int) int) (int, error) {
	reply := make(chan int, 1) // буфер: ответ не блокирует владельца
	err := s.do(ctx, func(m map[string]int) {
		m[key] = fn(m[key])
		reply <- m[key]
	})
	if err != nil {
		return 0, err
	}
	select {
	case v := <-reply:
		return v, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

// Snapshot возвращает копию всех счётчиков — изменение копии не влияет
// на хранилище. После Close — ErrClosed.
func (s *Store) Snapshot(ctx context.Context) (map[string]int, error) {
	reply := make(chan map[string]int, 1)
	err := s.do(ctx, func(m map[string]int) {
		reply <- maps.Clone(m) // копия делается в горутине-владельце
	})
	if err != nil {
		return nil, err
	}
	select {
	case m := <-reply:
		return m, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close останавливает горутину-владельца и ждёт её завершения.
// Повторный Close безопасен.
func (s *Store) Close() {
	s.once.Do(func() { close(s.quit) })
	<-s.done
}
