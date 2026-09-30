package main

import "errors"

var (
	ErrCycle   = errors.New("dependency cycle")
	ErrUnknown = errors.New("unknown dependency")
)

// Factory строит объект; get — получить зависимость по имени.
type Factory func(get func(name string) (any, error)) (any, error)

// Container — ленивый контейнер зависимостей. Не безопасен для горутин.
type Container struct {
	// ваши поля
}

func NewContainer() *Container {
	// ваш код
	return &Container{}
}

// Provide регистрирует (или заменяет) фабрику под именем name.
func (c *Container) Provide(name string, f Factory) {
	// ваш код
}

// Get возвращает объект name, при первом обращении создавая его фабрикой;
// дальше — тот же объект (фабрика вызывается один раз).
//   - Нет фабрики — ошибка, errors.Is(err, ErrUnknown), в тексте имя.
//   - Цикл — ошибка, errors.Is(err, ErrCycle), в тексте путь цикла вида
//     "a -> b -> a".
//   - Ошибки фабрик не кэшируются: следующий Get попробует снова.
func (c *Container) Get(name string) (any, error) {
	// ваш код
	return nil, nil
}
