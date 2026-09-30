package main

// change применяет изменение и возвращает изменение, которое его отменяет.
// Отмена отмены — снова повтор, поэтому undo и redo устроены одинаково.
type change func() change

// Store — хранилище ключ-значение с историей изменений.
// Не безопасно для горутин.
type Store struct {
	data  map[string]string
	undo  []change
	redo  []change
	batch *[]change // не nil — идёт Batch, изменения копятся здесь
}

func NewStore() *Store {
	return &Store{data: make(map[string]string)}
}

func (s *Store) Get(k string) (string, bool) {
	v, ok := s.data[k]
	return v, ok
}

// Set записывает значение. Это одно изменение в истории.
func (s *Store) Set(k, v string) {
	s.record(s.put(k, v, true))
}

// Delete удаляет ключ. Удаление отсутствующего ключа ничего не меняет и в
// историю не попадает.
func (s *Store) Delete(k string) {
	if _, ok := s.data[k]; !ok {
		return
	}
	s.record(s.put(k, "", false))
}

// Batch выполняет fn; все изменения внутри отменяются и повторяются одним
// Undo/Redo. Вложенный Batch сливается с внешним. Пустой Batch в историю не
// попадает. Если fn паникует, уже сделанные изменения остаются одной
// записью в истории, а паника летит дальше.
func (s *Store) Batch(fn func()) {
	if s.batch != nil {
		fn()
		return
	}
	var cs []change
	s.batch = &cs
	defer func() {
		// defer: даже после паники Store выходит из режима Batch.
		s.batch = nil
		if len(cs) > 0 {
			s.record(group(cs))
		}
	}()
	fn()
}

// Undo отменяет последнее изменение (или Batch). false — отменять нечего.
func (s *Store) Undo() bool {
	if len(s.undo) == 0 {
		return false
	}
	c := s.undo[len(s.undo)-1]
	s.undo = s.undo[:len(s.undo)-1]
	s.redo = append(s.redo, c())
	return true
}

// Redo повторяет последнее отменённое. Любое новое изменение очищает redo.
func (s *Store) Redo() bool {
	if len(s.redo) == 0 {
		return false
	}
	c := s.redo[len(s.redo)-1]
	s.redo = s.redo[:len(s.redo)-1]
	s.undo = append(s.undo, c())
	return true
}

// put меняет ключ и возвращает замыкание, которое вернёт прежнее состояние.
// Прежнее состояние — это и значение, и факт наличия ключа: отмена записи
// нового ключа должна его удалить, а не оставить пустую строку.
func (s *Store) put(k, v string, present bool) change {
	old, had := s.data[k]
	if present {
		s.data[k] = v
	} else {
		delete(s.data, k)
	}
	return func() change { return s.put(k, old, had) }
}

func (s *Store) record(c change) {
	if s.batch != nil {
		*s.batch = append(*s.batch, c)
		return
	}
	s.undo = append(s.undo, c)
	s.redo = nil
}

// group отменяет пачку в обратном порядке и возвращает пачку-обратную.
func group(cs []change) change {
	return func() change {
		inv := make([]change, 0, len(cs))
		for i := len(cs) - 1; i >= 0; i-- {
			inv = append(inv, cs[i]())
		}
		return group(inv)
	}
}
