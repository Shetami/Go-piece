package main

import "io"

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

type recorder struct {
	w   ResponseWriter
	rec *Recorder
}

// begin фиксирует статус 200, если заголовок ещё не отправлен явно.
func (r *recorder) begin() {
	if r.rec.Status == 0 {
		r.rec.Status = 200
	}
}

func (r *recorder) WriteHeader(code int) {
	if r.rec.Status != 0 {
		return // заголовок уже ушёл — повторный игнорируем и дальше не передаём
	}
	r.rec.Status = code
	r.w.WriteHeader(code)
}

func (r *recorder) Write(p []byte) (int, error) {
	r.begin()
	n, err := r.w.Write(p)
	r.rec.Bytes += int64(n)
	return n, err
}

func (r *recorder) Unwrap() ResponseWriter { return r.w }

func (r *recorder) flush() {
	r.begin()
	r.w.(Flusher).Flush()
}

func (r *recorder) readFrom(src io.Reader) (int64, error) {
	r.begin()
	n, err := r.w.(io.ReaderFrom).ReadFrom(src)
	r.rec.Bytes += n
	return n, err
}

// Опциональные возможности — отдельными типами, чтобы набор методов
// обёртки совпадал с набором методов оригинала. Каждый тип встраивает
// *recorder напрямую: встроить flusher и readerFrom вместе нельзя — Write
// пришёл бы с двух сторон на одной глубине, и селектор стал бы неоднозначным.
type flusher struct{ *recorder }

func (f flusher) Flush() { f.flush() }

type readerFrom struct{ *recorder }

func (rf readerFrom) ReadFrom(src io.Reader) (int64, error) { return rf.readFrom(src) }

type flusherReaderFrom struct{ *recorder }

func (x flusherReaderFrom) Flush()                                { x.flush() }
func (x flusherReaderFrom) ReadFrom(src io.Reader) (int64, error) { return x.readFrom(src) }

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
	r := &recorder{w: w, rec: rec}
	_, canFlush := w.(Flusher)
	_, canReadFrom := w.(io.ReaderFrom)
	switch {
	case canFlush && canReadFrom:
		return flusherReaderFrom{r}
	case canFlush:
		return flusher{r}
	case canReadFrom:
		return readerFrom{r}
	}
	return r
}
