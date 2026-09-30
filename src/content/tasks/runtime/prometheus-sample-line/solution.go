package main

import "strconv"

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
	// Копия для сортировки в массиве на стеке: пока он никуда не утекает,
	// escape-анализ оставляет его на стеке. Больше 16 меток — append
	// уйдёт в кучу, это допустимо.
	var tmp [16]Label
	ls := tmp[:0]
	for _, l := range labels {
		if l.Value == "" {
			continue
		}
		// Вставка в отсортированную позицию: меток единицы, O(n²) не страшен,
		// и не нужно передавать ls в обобщённую сортировку с замыканием.
		ls = append(ls, l)
		for i := len(ls) - 1; i > 0 && ls[i].Name < ls[i-1].Name; i-- {
			ls[i], ls[i-1] = ls[i-1], ls[i]
		}
	}

	dst = append(dst, name...)
	if len(ls) > 0 {
		dst = append(dst, '{')
		for i, l := range ls {
			if i > 0 {
				dst = append(dst, ',')
			}
			dst = append(dst, l.Name...)
			dst = append(dst, '=', '"')
			dst = appendEscaped(dst, l.Value)
			dst = append(dst, '"')
		}
		dst = append(dst, '}')
	}
	dst = append(dst, ' ')
	dst = strconv.AppendFloat(dst, value, 'g', -1, 64)
	return append(dst, '\n')
}

// appendEscaped — побайтово, без strings.Replacer и промежуточных строк.
func appendEscaped(dst []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\\':
			dst = append(dst, '\\', '\\')
		case '"':
			dst = append(dst, '\\', '"')
		case '\n':
			dst = append(dst, '\\', 'n')
		default:
			dst = append(dst, c)
		}
	}
	return dst
}
