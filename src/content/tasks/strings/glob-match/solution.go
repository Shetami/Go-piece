package main

import "errors"

// ErrBadPattern — шаблон заканчивается одиночной '\'.
var ErrBadPattern = errors.New("некорректный шаблон")

// Match сообщает, подходит ли name под шаблон pattern целиком.
//
//   - '*' — любая последовательность символов, в том числе пустая
//     (в отличие от path.Match, '/' тоже может входить);
//   - '?' — ровно один символ (руна, а не байт);
//   - '\' экранирует следующий символ: "\*" — буквальная звёздочка,
//     "\\" — обратная косая черта;
//   - остальные символы сравниваются как есть, с учётом регистра;
//   - '\' в самом конце шаблона — ErrBadPattern, даже если name заведомо
//     не подходит.
//
// Время — O(len(pattern)·len(name)) в худшем случае: никакого
// экспоненциального перебора на шаблонах вида "a*a*a*a*b".
func Match(pattern, name string) (bool, error) {
	// Сначала разбираем шаблон целиком: так ошибка находится всегда,
	// а не только если сравнение до неё дошло.
	type tok struct {
		r    rune
		kind byte // 'c' — символ, '?' — любой один, '*' — любая строка
	}
	var toks []tok
	pr := []rune(pattern)
	for i := 0; i < len(pr); i++ {
		switch pr[i] {
		case '\\':
			if i+1 == len(pr) {
				return false, ErrBadPattern
			}
			i++
			toks = append(toks, tok{pr[i], 'c'})
		case '*':
			toks = append(toks, tok{kind: '*'})
		case '?':
			toks = append(toks, tok{kind: '?'})
		default:
			toks = append(toks, tok{pr[i], 'c'})
		}
	}
	s := []rune(name)
	// Жадный разбор с одной точкой отката: запоминаем последнюю '*' и
	// позицию в name, с которой она начала «съедать». При несовпадении
	// звезда съедает ещё один символ. Более ранние звёзды пересматривать
	// не нужно — отсюда отсутствие экспоненты.
	p, i := 0, 0
	star, mark := -1, 0
	for i < len(s) {
		switch {
		case p < len(toks) && toks[p].kind == '*':
			star, mark = p, i
			p++
		case p < len(toks) && (toks[p].kind == '?' || toks[p].r == s[i]):
			p++
			i++
		case star >= 0:
			mark++
			p, i = star+1, mark
		default:
			return false, nil
		}
	}
	// name кончился — в шаблоне могут остаться только звёзды.
	for p < len(toks) && toks[p].kind == '*' {
		p++
	}
	return p == len(toks), nil
}
