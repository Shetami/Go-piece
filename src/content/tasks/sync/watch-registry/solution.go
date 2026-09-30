package main

import "sync"

type watcher struct {
	ch chan string // буфер 1: в нём лежит самое свежее непрочитанное значение
}

// Registry — реестр «ключ → значение» (адреса сервисов, фичефлаги) с
// подпиской на изменения. Get зовут очень часто, Set — редко.
type Registry struct {
	mu       sync.RWMutex
	values   map[string]string
	watchers map[string]map[*watcher]struct{}
}

func NewRegistry() *Registry {
	return &Registry{
		values:   make(map[string]string),
		watchers: make(map[string]map[*watcher]struct{}),
	}
}

// Get возвращает текущее значение ключа.
func (r *Registry) Get(key string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.values[key]
	return v, ok
}

// Set записывает значение и уведомляет подписчиков ключа. Set никогда не
// блокируется из-за медленного подписчика: если подписчик не успел
// прочитать прошлое уведомление, оно заменяется новым — подписчик
// может пропустить промежуточные значения, но последнее получит всегда.
func (r *Registry) Set(key, value string) {
	// Полная блокировка, а не RLock: два Set должны уведомлять по очереди,
	// иначе старое значение может лечь в канал поверх нового. И cancel не
	// закроет канал, пока мы в него пишем.
	r.mu.Lock()
	defer r.mu.Unlock()
	r.values[key] = value
	for w := range r.watchers[key] {
		offerLatest(w.ch, value)
	}
}

// offerLatest кладёт v в канал с буфером 1, выбрасывая устаревшее значение.
// Вызывается только под r.mu, поэтому других писателей в канал нет.
func offerLatest(ch chan string, v string) {
	select {
	case <-ch: // подписчик не успел прочитать — старое значение уже не нужно
	default:
	}
	ch <- v // место точно есть: писатель один, а читатель только освобождает
}

// Watch подписывается на изменения ключа. Если значение уже есть, оно
// сразу лежит в канале. cancel отписывает и закрывает канал; повторный
// вызов cancel ничего не делает, и его безопасно звать одновременно с Set.
func (r *Registry) Watch(key string) (updates <-chan string, cancel func()) {
	w := &watcher{ch: make(chan string, 1)}
	r.mu.Lock()
	if r.watchers[key] == nil {
		r.watchers[key] = make(map[*watcher]struct{})
	}
	r.watchers[key][w] = struct{}{}
	if v, ok := r.values[key]; ok {
		w.ch <- v
	}
	r.mu.Unlock()

	cancel = func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		// Закрываем, только если ещё подписан: повторный cancel — не паника.
		if _, ok := r.watchers[key][w]; !ok {
			return
		}
		delete(r.watchers[key], w)
		if len(r.watchers[key]) == 0 {
			delete(r.watchers, key) // не копим пустые множества
		}
		close(w.ch) // под той же блокировкой, что и отправка в Set
	}
	return w.ch, cancel
}

// Watchers — сколько сейчас активных подписок на ключ.
func (r *Registry) Watchers(key string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.watchers[key])
}
