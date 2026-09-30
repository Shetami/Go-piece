package main

import (
	"errors"
	"strings"
)

var ErrSyntax = errors.New("csv: syntax error")

// AppendFields разбирает одну строку CSV (разделитель ',', без перевода
// строки в конце) и дописывает поля к dst.
//
//   - поле в кавычках может содержать запятые и удвоенные кавычки "" → ";
//   - кавычка внутри поля без кавычек, незакрытая кавычка или символы
//     после закрывающей кавычки (кроме ',') — ErrSyntax (dst возвращается
//     как был);
//   - пустая строка — одно пустое поле.
//
// Горячий путь: поля без "" — подстроки line, без копирования. Если
// вместимости dst хватает и в строке нет "", функция не выделяет память;
// каждое поле с "" стоит ровно одно выделение.
func AppendFields(dst []string, line string) ([]string, error) {
	orig := len(dst)
	for {
		if !strings.HasPrefix(line, `"`) {
			// Поле без кавычек — до запятой; кавычка внутри запрещена.
			end := strings.IndexByte(line, ',')
			field := line
			if end >= 0 {
				field = line[:end]
			}
			if strings.IndexByte(field, '"') >= 0 {
				return dst[:orig], ErrSyntax
			}
			dst = append(dst, field) // подстрока — без копирования
			if end < 0 {
				return dst, nil
			}
			line = line[end+1:]
			continue
		}

		// Поле в кавычках: ищем закрывающую кавычку, пропуская "".
		body := line[1:]
		i, escaped := 0, false
		for {
			j := strings.IndexByte(body[i:], '"')
			if j < 0 {
				return dst[:orig], ErrSyntax // незакрытая кавычка
			}
			i += j
			if i+1 < len(body) && body[i+1] == '"' {
				escaped = true
				i += 2
				continue
			}
			break
		}
		raw := body[:i]
		if escaped {
			// Единственное выделение: Builder с точным Grow, String без копии.
			var b strings.Builder
			b.Grow(len(raw) - strings.Count(raw, `""`))
			for k := 0; k < len(raw); k++ {
				b.WriteByte(raw[k])
				if raw[k] == '"' {
					k++ // вторая кавычка из пары
				}
			}
			raw = b.String()
		}
		dst = append(dst, raw)

		rest := body[i+1:]
		switch {
		case rest == "":
			return dst, nil
		case rest[0] == ',':
			line = rest[1:]
		default:
			return dst[:orig], ErrSyntax // мусор после закрывающей кавычки
		}
	}
}
