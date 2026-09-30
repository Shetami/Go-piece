package main

import (
	"errors"
	"fmt"
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
	return func(c *Config) error {
		if d <= 0 {
			return fmt.Errorf("timeout %v must be positive: %w", d, ErrInvalid)
		}
		c.Timeout = d
		return nil
	}
}

// WithRetries задаёт число повторов: 0..10, иначе ошибка со словом "retries".
func WithRetries(n int) Option {
	return func(c *Config) error {
		if n < 0 || n > 10 {
			return fmt.Errorf("retries %d out of range 0..10: %w", n, ErrInvalid)
		}
		c.Retries = n
		return nil
	}
}

// WithTags добавляет теги к уже заданным; пустой тег — ошибка со словом "tag".
func WithTags(tags ...string) Option {
	// Своя копия: вызывающий может поменять слайс после WithTags(ts...).
	tags = append([]string(nil), tags...)
	return func(c *Config) error {
		for _, t := range tags {
			if t == "" {
				return fmt.Errorf("empty tag: %w", ErrInvalid)
			}
		}
		// Полное выражение c.Tags[:len:len] заставит append выделить новый
		// массив, даже если c.Tags делит память с чужим конфигом.
		c.Tags = append(c.Tags[:len(c.Tags):len(c.Tags)], tags...)
		return nil
	}
}

// New собирает конфиг: сначала значения по умолчанию (Timeout 5s,
// Retries 3, Tags nil), потом опции по порядку; nil-опции пропускаются.
// Пустой addr — ошибка со словом "addr". Ошибки всех опций и addr
// собираются в одну (errors.Join), каждая оборачивает ErrInvalid.
// При ошибке конфиг не возвращается (nil).
func New(addr string, opts ...Option) (*Config, error) {
	c := &Config{Addr: addr, Timeout: 5 * time.Second, Retries: 3}
	var errs []error
	if addr == "" {
		errs = append(errs, fmt.Errorf("addr is empty: %w", ErrInvalid))
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(c); err != nil {
			errs = append(errs, err)
		}
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return c, nil
}
