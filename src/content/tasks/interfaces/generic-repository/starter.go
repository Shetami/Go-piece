package main

import "errors"

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
	// ваши поля
}

func NewRepo[T Entity]() *Repo[T] {
	// ваш код
	return &Repo[T]{}
}

// Save сохраняет v под ключом v.Key().
//   - Если T реализует Validator и Validate вернул ошибку — вернуть ошибку,
//     для которой errors.Is(err, <ошибка Validate>), ничего не сохранять.
//   - Если T реализует Versioned: новая запись должна иметь версию 1,
//     обновление — ровно на 1 больше сохранённой; иначе ошибка
//     fmt.Errorf("repo: %q: %w", key, ErrConflict).
func (r *Repo[T]) Save(v T) error {
	// ваш код
	return nil
}

// Get возвращает сущность или нулевое T и fmt.Errorf("repo: %q: %w", key, ErrNotFound).
func (r *Repo[T]) Get(key string) (T, error) {
	// ваш код
	var zero T
	return zero, nil
}

// Delete удаляет сущность; нет такой — ошибка с ErrNotFound, как у Get.
func (r *Repo[T]) Delete(key string) error {
	// ваш код
	return nil
}

// List возвращает сущности, для которых keep вернул true (nil — все),
// по возрастанию Key. Пустой результат — пустой срез длины 0.
func (r *Repo[T]) List(keep func(T) bool) []T {
	// ваш код
	return nil
}
