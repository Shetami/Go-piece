package main

import (
	"context"
	"errors"
)

var ErrClosed = errors.New("store closed")

// Store — хранилище счётчиков, которым владеет одна горутина. Все
// операции уходят к ней запросами по каналу, ответ возвращается по каналу
// ответа из запроса. Мьютексов для данных нет — ими владеет горутина.
type Store struct {
	// ваши поля
}

// NewStore создаёт хранилище и запускает горутину-владельца.
func NewStore() *Store {
	// ваш код
	return &Store{}
}

// Update применяет fn к счётчику key (отсутствующий считается нулём)
// в горутине-владельце и возвращает новое значение. Если ctx отменён,
// пока запрос ждёт отправки или ответа, — ctx.Err() (изменение при этом
// может успеть примениться). Медленный или ушедший вызывающий не должен
// останавливать обработку чужих запросов. После Close — ErrClosed.
func (s *Store) Update(ctx context.Context, key string, fn func(old int) int) (int, error) {
	// ваш код
	return 0, nil
}

// Snapshot возвращает копию всех счётчиков — изменение копии не влияет
// на хранилище. После Close — ErrClosed.
func (s *Store) Snapshot(ctx context.Context) (map[string]int, error) {
	// ваш код
	return nil, nil
}

// Close останавливает горутину-владельца и ждёт её завершения.
// Повторный Close безопасен.
func (s *Store) Close() {
	// ваш код
}
