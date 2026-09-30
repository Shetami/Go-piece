package main

// cacheLine — размер строки кэша процессора (x86-64, большинство ARM64).
const cacheLine = 64

// shard — счётчик одного шарда. Шарды лежат в слайсе подряд, и два
// соседних шарда не должны попадать в одну строку кэша: иначе ядра,
// обновляющие разные шарды, всё равно мешают друг другу (false sharing).
// Требование: unsafe.Sizeof(shard{}) кратен cacheLine.
type shard struct {
	// ваши поля
}

// Counter — счётчик для горячего пути (число запросов, байт), который
// увеличивают из тысяч горутин. Число шардов — runtime.GOMAXPROCS(0) на
// момент создания. Add не выделяет память. Безопасен для конкурентного
// использования.
type Counter struct {
	// ваши поля
}

func NewCounter() *Counter {
	// ваш код
	return &Counter{}
}

func (c *Counter) Add(delta int64) {
	// ваш код
}

// Value — сумма по всем шардам.
func (c *Counter) Value() int64 {
	// ваш код
	return 0
}

// Shards — число шардов.
func (c *Counter) Shards() int {
	// ваш код
	return 0
}
