package main

import (
	"context"
	"errors"
)

// Plugin — минимум, который обязан уметь любой плагин.
type Plugin interface {
	Name() string
}

// Starter — опциональный интерфейс: плагин нужно запустить.
// (Опциональный io.Closer — плагин нужно остановить.)
type Starter interface {
	Start(ctx context.Context) error
}

// Factory создаёт плагин по конфигурации.
type Factory func(cfg map[string]string) (Plugin, error)

var (
	ErrDuplicate = errors.New("plugin already registered")
	ErrUnknown   = errors.New("unknown plugin")
)

// Registry — реестр фабрик, безопасный для параллельного использования.
type Registry struct {
	// ваши поля
}

func NewRegistry() *Registry {
	// ваш код
	return &Registry{}
}

// Register добавляет фабрику. Пустое имя или nil-фабрика — ошибка;
// повтор имени — ошибка с ErrDuplicate и именем в тексте.
func (r *Registry) Register(name string, f Factory) error {
	// ваш код
	return nil
}

// Names — зарегистрированные имена по возрастанию.
func (r *Registry) Names() []string {
	// ваш код
	return nil
}

// Set — загруженные и запущенные плагины.
type Set struct {
	// ваши поля
}

// Plugins — плагины в порядке загрузки.
func (s *Set) Plugins() []Plugin {
	// ваш код
	return nil
}

// Close закрывает плагины, реализующие io.Closer, в обратном порядке
// загрузки; ошибки — через errors.Join. Повторный Close — nil.
func (s *Set) Close() error {
	// ваш код
	return nil
}

// Load создаёт плагины по именам в заданном порядке (cfg[name] — конфигурация)
// и запускает тех, кто реализует Starter.
//   - Неизвестное имя — ошибка с ErrUnknown, и ни одна фабрика не вызывается.
//   - Перед каждым Start проверяется ctx: отменён — это ошибка запуска.
//   - Ошибка фабрики или запуска: все уже созданные плагины, включая
//     упавший на Start, закрываются (io.Closer) в обратном порядке, а Load
//     возвращает errors.Join(исходная ошибка, ошибки закрытия).
func (r *Registry) Load(ctx context.Context, names []string, cfg map[string]map[string]string) (*Set, error) {
	// ваш код
	return &Set{}, nil
}
