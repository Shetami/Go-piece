package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var (
	ErrCycle   = errors.New("dependency cycle")
	ErrUnknown = errors.New("unknown dependency")
)

// Factory строит объект; get — получить зависимость по имени.
type Factory func(get func(name string) (any, error)) (any, error)

// Container — ленивый контейнер зависимостей. Не безопасен для горутин.
type Container struct {
	factories map[string]Factory
	built     map[string]any
}

func NewContainer() *Container {
	return &Container{factories: make(map[string]Factory), built: make(map[string]any)}
}

// Provide регистрирует (или заменяет) фабрику под именем name.
func (c *Container) Provide(name string, f Factory) {
	c.factories[name] = f
}

// Get возвращает объект name, при первом обращении создавая его фабрикой;
// дальше — тот же объект (фабрика вызывается один раз).
//   - Нет фабрики — ошибка, errors.Is(err, ErrUnknown), в тексте имя.
//   - Цикл — ошибка, errors.Is(err, ErrCycle), в тексте путь цикла вида
//     "a -> b -> a".
//   - Ошибки фабрик не кэшируются: следующий Get попробует снова.
func (c *Container) Get(name string) (any, error) {
	return c.get(name, nil)
}

// path — цепочка имён, которые строятся прямо сейчас выше по стеку.
func (c *Container) get(name string, path []string) (any, error) {
	if v, ok := c.built[name]; ok {
		return v, nil
	}
	// Цикл — это имя, которое уже строится на текущем пути, а не любое
	// встреченное: ромб a→b→d, a→c→d циклом не является.
	if i := slices.Index(path, name); i >= 0 {
		loop := append(slices.Clone(path[i:]), name)
		return nil, fmt.Errorf("%w: %s", ErrCycle, strings.Join(loop, " -> "))
	}
	f, ok := c.factories[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknown, name)
	}
	// Полное выражение среза: соседние зависимости не пишут в общий массив.
	path = append(path[:len(path):len(path)], name)
	v, err := f(func(dep string) (any, error) {
		return c.get(dep, path)
	})
	if err != nil {
		return nil, fmt.Errorf("build %s: %w", name, err)
	}
	c.built[name] = v
	return v, nil
}
