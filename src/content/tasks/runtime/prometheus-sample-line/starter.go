package main

// Label — метка метрики.
type Label struct {
	Name, Value string
}

// AppendSample дописывает к dst строку текстового формата Prometheus:
//
//	name{a="1",b="x y"} 0.5\n
//
//   - метки выводятся отсортированными по Name; слайс labels не меняется;
//   - метки с пустым Value пропускаются (в Prometheus это то же, что
//     отсутствие метки); если меток не осталось — без фигурных скобок;
//   - в значении метки экранируются \ → \\, " → \", перевод строки → \n;
//   - число — как strconv.FormatFloat(v, 'g', -1, 64): 1e+06, NaN, +Inf, -Inf.
//
// Горячий путь /metrics: при достаточной вместимости dst и не больше
// 16 меток функция не выделяет память.
func AppendSample(dst []byte, name string, labels []Label, value float64) []byte {
	// ваш код
	return dst
}
