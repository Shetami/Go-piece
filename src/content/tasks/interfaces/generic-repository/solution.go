package main

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"sync"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("version conflict")
)

// Entity — то, что умеет хранить репозиторий.
type Entity interface {
	Key() string
}

// Validator — опциональный интерфейс: сущность умеет себя проверять.
type Validator interface {
	Validate() error
}

// Versioned — опциональный интерфейс для оптимистичной блокировки.
type Versioned interface {
	Version() int
}

// Repo — хранилище сущностей в памяти, безопасное для параллельного доступа.
type Repo[T Entity] struct {
	mu sync.RWMutex
	m  map[string]T
}

func NewRepo[T Entity]() *Repo[T] {
	return &Repo[T]{m: make(map[string]T)}
}

// Save сохраняет v под ключом v.Key().
//   - Если T реализует Validator и Validate вернул ошибку — вернуть ошибку,
//     для которой errors.Is(err, <ошибка Validate>), ничего не сохранять.
//   - Если T реализует Versioned: новая запись должна иметь версию 1,
//     обновление — ровно на 1 больше сохранённой; иначе ошибка
//     fmt.Errorf("repo: %q: %w", key, ErrConflict).
func (r *Repo[T]) Save(v T) error {
	// К параметру типа нельзя применить утверждение типа напрямую — только через any.
	if val, ok := any(v).(Validator); ok {
		if err := val.Validate(); err != nil {
			return fmt.Errorf("repo: invalid %q: %w", v.Key(), err)
		}
	}
	key := v.Key()
	r.mu.Lock()
	defer r.mu.Unlock()
	if nv, ok := any(v).(Versioned); ok {
		want := 1
		if old, exists := r.m[key]; exists {
			want = any(old).(Versioned).Version() + 1
		}
		if nv.Version() != want {
			return fmt.Errorf("repo: %q: %w", key, ErrConflict)
		}
	}
	r.m[key] = v
	return nil
}

// Get возвращает сущность или нулевое T и fmt.Errorf("repo: %q: %w", key, ErrNotFound).
func (r *Repo[T]) Get(key string) (T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.m[key]
	if !ok {
		var zero T
		return zero, fmt.Errorf("repo: %q: %w", key, ErrNotFound)
	}
	return v, nil
}

// Delete удаляет сущность; нет такой — ошибка с ErrNotFound, как у Get.
func (r *Repo[T]) Delete(key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.m[key]; !ok {
		return fmt.Errorf("repo: %q: %w", key, ErrNotFound)
	}
	delete(r.m, key)
	return nil
}

// List возвращает сущности, для которых keep вернул true (nil — все),
// по возрастанию Key. Пустой результат — пустой срез длины 0.
func (r *Repo[T]) List(keep func(T) bool) []T {
	r.mu.RLock()
	out := make([]T, 0, len(r.m))
	for _, v := range r.m {
		if keep == nil || keep(v) {
			out = append(out, v)
		}
	}
	r.mu.RUnlock()
	// Порядок обхода мапы случаен — сортируем явно.
	slices.SortFunc(out, func(a, b T) int { return cmp.Compare(a.Key(), b.Key()) })
	return out
}
