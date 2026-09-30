package main

import (
	"strings"
	"unicode"
)

// translit — транслитерация строчных русских букв.
var translit = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo",
	'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
	'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
	'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shch",
	'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
}

// Slug делает из заголовка кусок URL: "Щука и ёж: 2 истории!" → "shchuka-i-yozh-2-istorii".
//
//   - регистр не важен, результат в нижнем регистре;
//   - русские буквы — по таблице translit (заглавные тоже: "Щ" → "shch");
//     'ъ' и 'ь' выпадают, не разрывая слово: "подъезд" → "podezd";
//   - латинские буквы a–z и цифры 0–9 остаются;
//   - прочие буквы (é, ü, 日…) выбрасываются, тоже не разрывая слово;
//   - всё остальное (пробелы, знаки, дефисы, подчёркивания) — разделитель;
//     подряд идущие разделители дают один '-', по краям '-' нет;
//   - maxLen > 0 ограничивает длину результата в байтах: режем по границе
//     слова (последний '-', который помещается), а если первое слово
//     длиннее maxLen — режем само слово. maxLen <= 0 — без ограничения.
func Slug(title string, maxLen int) string {
	var b strings.Builder
	sep := false // был разделитель после последнего слова
	for _, r := range title {
		r = unicode.ToLower(r)
		var out string
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			out = string(r)
		case translit[r] != "" || r == 'ъ' || r == 'ь':
			out = translit[r]
		case unicode.IsLetter(r) || unicode.Is(unicode.Mn, r):
			continue // незнакомая буква: выбрасываем, слово не разрываем
		default:
			sep = true
			continue
		}
		if out == "" {
			continue
		}
		// Дефис ставим лениво — только перед следующим словом: так не
		// бывает ни дефисов по краям, ни двойных.
		if sep && b.Len() > 0 {
			b.WriteByte('-')
		}
		sep = false
		b.WriteString(out)
	}
	s := b.String()
	if maxLen > 0 && len(s) > maxLen {
		cut := s[:maxLen] // в слаге только ASCII — резать по байтам безопасно
		if s[maxLen] == '-' {
			return cut // граница как раз между словами
		}
		if i := strings.LastIndexByte(cut, '-'); i > 0 {
			return cut[:i]
		}
		return cut
	}
	return s
}
