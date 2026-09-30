package main

import (
	"maps"
	"sync"
)

// Registry — потокобезопасный реестр настроек: раздел → ключ → значение.
//
//   - Set пишет значение, создавая раздел при необходимости.
//   - Get читает; ok=false, если нет раздела или ключа.
//   - Delete удаляет ключ; раздел, в котором не осталось ключей, исчезает.
//   - Snapshot возвращает независимую копию всего реестра: правки снимка
//     (в том числе внутри разделов) не видны реестру, а правки реестра —
//     снимку. У пустого реестра снимок — пустая, но не nil мапа.
type Registry struct {
	mu   sync.RWMutex
	data map[string]map[string]string
}

func NewRegistry() *Registry {
	return &Registry{data: make(map[string]map[string]string)}
}

func (r *Registry) Set(section, key, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	sec := r.data[section]
	if sec == nil {
		sec = make(map[string]string) // запись в nil-мапу — паника
		r.data[section] = sec
	}
	sec[key] = value
}

func (r *Registry) Get(section, key string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.data[section][key] // чтение из nil-мапы безопасно
	return v, ok
}

func (r *Registry) Delete(section, key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	sec := r.data[section]
	delete(sec, key)
	if len(sec) == 0 {
		delete(r.data, section)
	}
}

func (r *Registry) Snapshot() map[string]map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	snap := make(map[string]map[string]string, len(r.data))
	for name, sec := range r.data {
		snap[name] = maps.Clone(sec) // копируем и каждый раздел, не только внешнюю мапу
	}
	return snap
}
