package main

import "bytes"

// Record — одна строка CSV, разрезанная на поля.
type Record [][]byte

// ReadAll читает строки через next, пока тот не вернёт ok == false, и режет
// каждую по разделителю sep. Кавычек нет: sep внутри поля не встречается.
//
// next ведёт себя как bufio.Scanner.Bytes: отдаёт буфер, который будет
// перезаписан при следующем вызове, — поля не должны на него ссылаться.
// Пустые строки пропускаются. "a,,b," — четыре поля, второе и последнее пустые.
// На каждую запись — одна копия байт строки; поля — подслайсы этой копии,
// и append к полю не затирает следующее поле.
func ReadAll(next func() (line []byte, ok bool), sep byte) []Record {
	var out []Record
	for {
		line, ok := next()
		if !ok {
			return out
		}
		if len(line) == 0 {
			continue
		}
		own := bytes.Clone(line) // единственная копия: буфер next скоро перезапишут
		rec := make(Record, 0, bytes.Count(own, []byte{sep})+1)
		for {
			i := bytes.IndexByte(own, sep)
			if i < 0 {
				rec = append(rec, own[:len(own):len(own)])
				break
			}
			// Третий индекс: append к полю не залезет в соседнее.
			rec = append(rec, own[:i:i])
			own = own[i+1:]
		}
		out = append(out, rec)
	}
}
