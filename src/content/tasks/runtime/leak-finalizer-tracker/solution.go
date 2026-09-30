package main

import (
	"errors"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
)

var ErrClosed = errors.New("resource already closed")

// Tracker выдаёт ресурсы (условные соединения) и ловит забытые Close:
// ресурс, ставший недостижимым без Close, после сборки мусора попадает
// в список утёкших — так os.File и net.Conn закрывают за собой
// дескрипторы. Безопасен для конкурентного использования.
// Сам Tracker не должен мешать сборщику мусора собирать ресурсы.
type Tracker struct {
	open   atomic.Int64 // счётчик, а не множество *Resource: множество держало бы их живыми
	mu     sync.Mutex
	leaked []string
}

// Resource — выданный ресурс.
type Resource struct {
	name   string
	t      *Tracker
	closed atomic.Bool
}

// Open выдаёт новый ресурс с именем name.
func (t *Tracker) Open(name string) *Resource {
	r := &Resource{name: name, t: t}
	t.open.Add(1)
	// Финализатор получает объект параметром. Замыкание, захватившее r,
	// сделало бы r достижимым из самого финализатора — и он бы не сработал.
	runtime.SetFinalizer(r, (*Resource).onCollected)
	return r
}

func (r *Resource) onCollected() {
	r.t.open.Add(-1)
	r.t.mu.Lock() // финализаторы выполняются в отдельной горутине
	r.t.leaked = append(r.t.leaked, r.name)
	r.t.mu.Unlock()
}

// Close закрывает ресурс; повторный Close возвращает ErrClosed.
// Закрытый ресурс утёкшим не считается.
func (r *Resource) Close() error {
	if !r.closed.CompareAndSwap(false, true) {
		return ErrClosed
	}
	// Снимаем финализатор: закрытый ресурс — не утечка, и объект
	// освободится за один цикл GC, а не за два.
	runtime.SetFinalizer(r, nil)
	r.t.open.Add(-1)
	return nil
}

// Leaked — имена утёкших ресурсов (собранных без Close), по возрастанию.
func (t *Tracker) Leaked() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := slices.Clone(t.leaked)
	slices.Sort(out)
	return out
}

// OpenCount — сколько ресурсов открыто сейчас: выданы, не закрыты и не утекли.
func (t *Tracker) OpenCount() int { return int(t.open.Load()) }
