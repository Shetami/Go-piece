package main

import "sync"

// Broker рассылает сообщения всем подписчикам.
// Все методы безопасны для вызова из разных горутин.
type Broker struct {
	mu   sync.Mutex
	subs map[chan string]struct{}
}

func NewBroker() *Broker {
	return &Broker{subs: make(map[chan string]struct{})}
}

// Subscribe заводит подписчика с буфером на buf сообщений.
// Возвращает канал сообщений и функцию отписки. Отписка закрывает канал;
// повторный вызов отписки ничего не делает.
func (b *Broker) Subscribe(buf int) (<-chan string, func()) {
	ch := make(chan string, buf)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			// Под тем же мьютексом, что и Publish: иначе Publish мог бы
			// отправить в канал, который мы только что закрыли.
			b.mu.Lock()
			delete(b.subs, ch)
			close(ch)
			b.mu.Unlock()
		})
	}
	return ch, unsubscribe
}

// Publish отправляет msg каждому подписчику, не блокируясь: если буфер
// подписчика полон, сообщение для него выбрасывается.
// Возвращает, скольким подписчикам сообщение доставлено.
func (b *Broker) Publish(msg string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	delivered := 0
	for ch := range b.subs {
		select {
		case ch <- msg:
			delivered++
		default: // медленный подписчик не должен тормозить остальных
		}
	}
	return delivered
}
