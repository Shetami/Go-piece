package main

import (
	"errors"
	"sync"
)

var (
	ErrNotFound = errors.New("не найдено")
	ErrExists   = errors.New("уже существует")
	ErrConflict = errors.New("конфликт версий")
)

// VersionError — подробности конфликта оптимистической блокировки.
// Текст: "user <ID>: версия <Want>, в базе <Have>";
// errors.Is(err, ErrConflict) должен быть истинным.
type VersionError struct {
	ID         string
	Want, Have int
}

type User struct {
	ID      string
	Name    string
	Balance int
	Version int
}

// Repo — хранилище в памяти, безопасное для конкурентного доступа.
type Repo struct {
	mu sync.Mutex
	m  map[string]User
}

func NewRepo() *Repo { return &Repo{m: make(map[string]User)} }

func (e *VersionError) Error() string {
	// ваш код
	return ""
}

func (e *VersionError) Is(target error) bool {
	// ваш код
	return false
}

// Get возвращает пользователя. Нет такого — ошибка, для которой
// errors.Is(err, ErrNotFound), с ID в тексте.
func (r *Repo) Get(id string) (User, error) {
	// ваш код
	return User{}, nil
}

// Create добавляет пользователя с Version = 1. ID занят — ошибка с
// ErrExists (errors.Is) и ID в тексте.
func (r *Repo) Create(u User) error {
	// ваш код
	return nil
}

// Update сохраняет u, если u.Version совпадает с хранимой, и возвращает
// сохранённую запись с Version+1. Нет записи — ErrNotFound; версия не
// совпала — *VersionError.
func (r *Repo) Update(u User) (User, error) {
	// ваш код
	return u, nil
}

// Modify — read-modify-write с повтором на конфликте: прочитать запись,
// применить fn к копии, сохранить через Update. Конфликт версий — начать
// заново (до 50 попыток, потом вернуть последнюю ошибку конфликта).
// Ошибка fn возвращается как есть, и Update тогда не вызывается;
// ErrNotFound — тоже сразу, без повторов.
func Modify(r *Repo, id string, fn func(*User) error) error {
	// ваш код
	return nil
}
