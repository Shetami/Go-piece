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
	return e.Resource + " " + e.ID + ": " + ErrNotFound.Error()
}

// Is связывает тип с сентинелом: errors.Is спрашивает каждое звено цепочки.
func (e *NotFoundError) Is(target error) bool {
	return target == ErrNotFound
}
