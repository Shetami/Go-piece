package main

// Batches раскладывает сообщения по пачкам, сохраняя порядок. В пачке не
// больше maxCount сообщений и не больше maxBytes байт суммарно. Сообщение
// длиннее maxBytes уходит отдельной пачкой из одного элемента. Пустых пачек
// не бывает. Пачки — подслайсы msgs без копирования, и append к пачке не
// портит соседнюю. maxCount <= 0 или maxBytes <= 0 — паника.
func Batches(msgs [][]byte, maxCount, maxBytes int) [][][]byte {
	if maxCount <= 0 || maxBytes <= 0 {
		panic("Batches: лимиты должны быть положительными")
	}
	var out [][][]byte
	start, size := 0, 0 // текущая пачка — msgs[start:i], в ней size байт
	for i, m := range msgs {
		// Закрываем пачку, только если она не пуста: так огромное
		// сообщение не порождает пустую пачку перед собой.
		if i > start && (i-start == maxCount || size+len(m) > maxBytes) {
			out = append(out, msgs[start:i:i])
			start, size = i, 0
		}
		size += len(m)
	}
	if start < len(msgs) {
		out = append(out, msgs[start:len(msgs):len(msgs)])
	}
	return out
}
