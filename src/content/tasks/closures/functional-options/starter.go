package main

import (
	"errors"
	"time"
)

// ErrInvalid оборачивают все ошибки конфигурации.
var ErrInvalid = errors.New("invalid config")

// Config — настройки клиента.
type Config struct {
	Addr    string
	Timeout time.Duration
	Retries int
	Tags    []string
}

// Option меняет конфиг или сообщает, почему не может.
type Option func(*Config) error

// WithTimeout задаёт таймаут; d должен быть > 0, иначе ошибка со словом "timeout".
func WithTimeout(d time.Duration) Option {
	// ваш код
	return func(c *Config) error { return nil }
}

// WithRetries задаёт число повторов: 0..10, иначе ошибка со словом "retries".
func WithRetries(n int) Option {
	// ваш код
	return func(c *Config) error { return nil }
}

// WithTags добавляет теги к уже заданным; пустой тег — ошибка со словом "tag".
func WithTags(tags ...string) Option {
	// ваш код
	return func(c *Config) error { return nil }
}

// New собирает конфиг: сначала значения по умолчанию (Timeout 5s,
// Retries 3, Tags nil), потом опции по порядку; nil-опции пропускаются.
// Пустой addr — ошибка со словом "addr". Ошибки всех опций и addr
// собираются в одну (errors.Join), каждая оборачивает ErrInvalid.
// При ошибке конфиг не возвращается (nil).
func New(addr string, opts ...Option) (*Config, error) {
	// ваш код
	return nil, nil
}
