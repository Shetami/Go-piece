package main

// Tee копирует каждое значение из in в оба выходных канала.
// Оба выхода закрываются, когда закрылся in.
func Tee(in <-chan int) (<-chan int, <-chan int) {
	out1, out2 := make(chan int), make(chan int)
	go func() {
		defer close(out1)
		defer close(out2)
		for v := range in {
			// Отдаём значение обоим в любом порядке: кто первым готов читать.
			// Отправивший канал обнуляем — nil выключает его ветку select.
			o1, o2 := out1, out2
			for range 2 {
				select {
				case o1 <- v:
					o1 = nil
				case o2 <- v:
					o2 = nil
				}
			}
		}
	}()
	return out1, out2
}
