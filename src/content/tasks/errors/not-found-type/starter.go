package main

import "errors"

var ErrNotFound = errors.New("не найдено")

// NotFoundError — «не найдено» с подробностями.
// Error() возвращает "<Resource> <ID>: не найдено",
// а errors.Is(err, ErrNotFound) для него истинно.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	// ваш код
	return ""
}
