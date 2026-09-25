package main

// Split раздаёт значения из in по n выходам по кругу: первое — в выход 0,
// второе — в выход 1, ..., n-е — снова в выход 0. Когда in закрыт,
// все выходы закрываются. Для n <= 0 — паника.
func Split[T any](in <-chan T, n int) []<-chan T {
	if n <= 0 {
		panic("Split: n должно быть больше нуля")
	}
	outs := make([]chan T, n)
	// Наружу — только каналы на чтение: закрывать и писать их может лишь Split.
	ro := make([]<-chan T, n)
	for i := range outs {
		outs[i] = make(chan T)
		ro[i] = outs[i]
	}
	go func() {
		// Одна горутина пишет во все выходы — она же единственная их и закрывает.
		defer func() {
			for _, ch := range outs {
				close(ch)
			}
		}()
		i := 0
		for v := range in {
			outs[i] <- v
			i = (i + 1) % n
		}
	}()
	return ro
}
