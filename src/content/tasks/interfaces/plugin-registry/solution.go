package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"sync"
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
	mu        sync.RWMutex
	factories map[string]Factory
}

func NewRegistry() *Registry {
	return &Registry{factories: map[string]Factory{}}
}

// Register добавляет фабрику. Пустое имя или nil-фабрика — ошибка;
// повтор имени — ошибка с ErrDuplicate и именем в тексте.
func (r *Registry) Register(name string, f Factory) error {
	if name == "" || f == nil {
		return errors.New("register: empty name or nil factory")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.factories[name]; ok {
		return fmt.Errorf("register %q: %w", name, ErrDuplicate)
	}
	r.factories[name] = f
	return nil
}

// Names — зарегистрированные имена по возрастанию.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.factories))
	for n := range r.factories {
		names = append(names, n)
	}
	slices.Sort(names)
	return names
}

// Set — загруженные и запущенные плагины.
type Set struct {
	plugins []Plugin
	once    sync.Once
}

// Plugins — плагины в порядке загрузки.
func (s *Set) Plugins() []Plugin { return s.plugins }

// Close закрывает плагины, реализующие io.Closer, в обратном порядке
// загрузки; ошибки — через errors.Join. Повторный Close — nil.
func (s *Set) Close() error {
	var err error
	s.once.Do(func() { err = closeAll(s.plugins) })
	return err
}

func closeAll(ps []Plugin) error {
	var errs []error
	for _, p := range slices.Backward(ps) {
		if c, ok := p.(io.Closer); ok {
			if err := c.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close %s: %w", p.Name(), err))
			}
		}
	}
	return errors.Join(errs...)
}

// Load создаёт плагины по именам в заданном порядке (cfg[name] — конфигурация)
// и запускает тех, кто реализует Starter.
//   - Неизвестное имя — ошибка с ErrUnknown, и ни одна фабрика не вызывается.
//   - Перед каждым Start проверяется ctx: отменён — это ошибка запуска.
//   - Ошибка фабрики или запуска: все уже созданные плагины, включая
//     упавший на Start, закрываются (io.Closer) в обратном порядке, а Load
//     возвращает errors.Join(исходная ошибка, ошибки закрытия).
func (r *Registry) Load(ctx context.Context, names []string, cfg map[string]map[string]string) (*Set, error) {
	r.mu.RLock()
	fs := make([]Factory, len(names))
	for i, n := range names {
		fs[i] = r.factories[n]
	}
	r.mu.RUnlock()
	// Сначала проверяем все имена: создавать что-то до полной проверки незачем.
	for i, f := range fs {
		if f == nil {
			return nil, fmt.Errorf("load %q: %w", names[i], ErrUnknown)
		}
	}

	var created []Plugin
	fail := func(err error) (*Set, error) {
		return nil, errors.Join(err, closeAll(created))
	}
	for i, f := range fs {
		p, err := f(cfg[names[i]])
		if err != nil {
			return fail(fmt.Errorf("create %s: %w", names[i], err))
		}
		created = append(created, p) // создан — значит, его надо будет закрыть
		if s, ok := p.(Starter); ok {
			if err := ctx.Err(); err != nil {
				return fail(fmt.Errorf("start %s: %w", names[i], err))
			}
			if err := s.Start(ctx); err != nil {
				return fail(fmt.Errorf("start %s: %w", names[i], err))
			}
		}
	}
	return &Set{plugins: created}, nil
}
