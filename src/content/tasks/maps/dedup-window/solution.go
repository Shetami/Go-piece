package main

import "time"

// Dedup отбрасывает повторы событий в скользящем окне.
//
//   - Seen(id, at) возвращает true, если событие id уже было принято меньше
//     чем window назад (at - принятое < window), — это повтор.
//     Иначе событие принимается (false), и окно для id отсчитывается от at.
//   - Повторы окно не продлевают: оно считается от последнего ПРИНЯТОГО
//     события, а не от последнего увиденного.
//   - at не убывает от вызова к вызову.
//   - Len — сколько id сейчас хранится. Записи, чьё окно истекло к моменту
//     последнего вызова Seen, храниться не должны: при потоке уникальных id
//     память не растёт бесконечно. Чистка — амортизированно O(1) на вызов,
//     без обхода всей мапы.
type Dedup struct {
	window   time.Duration
	accepted map[string]time.Time // id → время последнего принятого события
	queue    []dedupRec           // принятые события в порядке времени
}

type dedupRec struct {
	id string
	at time.Time
}

func NewDedup(window time.Duration) *Dedup {
	return &Dedup{window: window, accepted: make(map[string]time.Time)}
}

func (d *Dedup) evict(now time.Time) {
	// at не убывает, значит, очередь отсортирована: истёкшие — в начале.
	i := 0
	for ; i < len(d.queue) && now.Sub(d.queue[i].at) >= d.window; i++ {
		r := d.queue[i]
		// Запись могла устареть: id с тех пор приняли заново.
		// Удаляем из мапы, только если там именно это время.
		if t, ok := d.accepted[r.id]; ok && t.Equal(r.at) {
			delete(d.accepted, r.id)
		}
	}
	d.queue = d.queue[i:]
	// Сжимаем, когда голова «съела» больше половины массива: иначе
	// d.queue[i:] держит весь старый массив в памяти.
	if cap(d.queue) > 64 && len(d.queue) < cap(d.queue)/2 {
		d.queue = append([]dedupRec(nil), d.queue...)
	}
}

func (d *Dedup) Seen(id string, at time.Time) bool {
	d.evict(at)
	if t, ok := d.accepted[id]; ok && at.Sub(t) < d.window {
		return true // повтор; время принятия не трогаем
	}
	d.accepted[id] = at
	d.queue = append(d.queue, dedupRec{id, at})
	return false
}

func (d *Dedup) Len() int { return len(d.accepted) }
