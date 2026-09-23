package main

import "sync"

// Group склеивает одновременные вызовы с одинаковым ключом:
// пока fn для ключа выполняется, остальные вызовы с тем же ключом
// ждут её результата, а не запускают свою. Это не кэш: вызов после
// завершения снова выполняет fn.
type Group struct {
	mu    sync.Mutex
	calls map[string]*call
}

// call — один выполняющийся вызов. Ждущие ждут на wg,
// результат читают после Wait.
type call struct {
	wg  sync.WaitGroup
	val int
	err error
}

func (g *Group) Do(key string, fn func() (int, error)) (int, error) {
	g.mu.Lock()
	if g.calls == nil {
		g.calls = make(map[string]*call)
	}
	if c, ok := g.calls[key]; ok {
		// Вызов уже идёт — отпускаем мьютекс и ждём его результата.
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := &call{}
	c.wg.Add(1)
	g.calls[key] = c
	g.mu.Unlock()

	// fn выполняется без мьютекса: иначе вызовы с разными ключами ждали бы друг друга.
	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.calls, key) // следующий вызов после завершения — снова fn
	g.mu.Unlock()
	return c.val, c.err
}
