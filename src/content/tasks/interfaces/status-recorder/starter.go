package main

// ResponseWriter — упрощённый http.ResponseWriter.
type ResponseWriter interface {
	Write(p []byte) (int, error)
	WriteHeader(code int)
}

// Flusher — опциональный интерфейс: отправить буфер клиенту немедленно.
type Flusher interface {
	Flush()
}

// Recorder — то, что обёртка узнала об ответе.
type Recorder struct {
	Status int   // итоговый статус; 0 — ответ ещё не начат
	Bytes  int64 // сколько байт тела реально записано
}

// Wrap возвращает обёртку над w, которая пишет в rec статус и размер тела.
//   - Первый WriteHeader(code) задаёт статус и передаётся в w; все
//     последующие игнорируются и в w не передаются.
//   - Write, ReadFrom или Flush до WriteHeader начинают ответ со статусом 200;
//     в w при этом WriteHeader не вызывается (w сделает это сам), а
//     поздний WriteHeader уже игнорируется.
//   - Bytes — сумма реально записанных байт (n из Write и ReadFrom).
//   - Обёртка реализует Flusher тогда и только тогда, когда w реализует
//     Flusher; то же для io.ReaderFrom. Flush и ReadFrom идут в w.
//   - У обёртки есть метод Unwrap() ResponseWriter, возвращающий w.
func Wrap(w ResponseWriter, rec *Recorder) ResponseWriter {
	// ваш код
	return w
}
