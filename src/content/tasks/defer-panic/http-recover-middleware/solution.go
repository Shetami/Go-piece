package main

import (
	"fmt"
	"net/http"
)

// Recover — middleware, которое не даёт панике обработчика оборвать соединение
// молча.
//   - Паника, пока обработчик ещё ничего не записал в ответ: ответ 500 с
//     телом "internal error" и одна запись в logf, где есть значение паники,
//     метод и путь запроса.
//   - Паника после того, как обработчик уже начал ответ (WriteHeader или
//     Write): ответ не трогать — дописывать в него нельзя, — только logf.
//   - panic(http.ErrAbortHandler) — штатный способ оборвать ответ: не
//     логировать и паниковать дальше тем же значением.
//   - Без паники поведение обработчика не меняется.
func Recover(logf func(msg string), next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tw := &trackingWriter{ResponseWriter: w}
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if rec == http.ErrAbortHandler {
				panic(rec) // сигнал для net/http, не ошибка
			}
			logf(fmt.Sprintf("паника в %s %s: %v", r.Method, r.URL.Path, rec))
			if !tw.started {
				http.Error(tw, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(tw, r)
	})
}

// trackingWriter запоминает, начат ли уже ответ.
type trackingWriter struct {
	http.ResponseWriter
	started bool
}

func (w *trackingWriter) WriteHeader(code int) {
	w.started = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *trackingWriter) Write(b []byte) (int, error) {
	w.started = true // первый Write неявно отправляет 200
	return w.ResponseWriter.Write(b)
}
