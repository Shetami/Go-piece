package main

// Batches раскладывает сообщения по пачкам, сохраняя порядок. В пачке не
// больше maxCount сообщений и не больше maxBytes байт суммарно. Сообщение
// длиннее maxBytes уходит отдельной пачкой из одного элемента. Пустых пачек
// не бывает. Пачки — подслайсы msgs без копирования, и append к пачке не
// портит соседнюю. maxCount <= 0 или maxBytes <= 0 — паника.
func Batches(msgs [][]byte, maxCount, maxBytes int) [][][]byte {
	// ваш код
	return nil
}
