package main

import "strings"

// NormalizeText приводит пробелы в тексте к норме:
//   - внутри абзаца любые пробельные символы (пробелы, табы, неразрывный
//     пробел U+00A0, одиночные переводы строк) схлопываются в один пробел;
//   - абзацы разделены пустыми строками — строками, где нет ничего, кроме
//     пробельных символов; в результате между абзацами ровно "\n\n";
//   - по краям текста и абзацев пробелов нет;
//   - окончания строк бывают и "\n", и "\r\n".
func NormalizeText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	var paras []string
	var cur []string // слова текущего абзаца
	flush := func() {
		if len(cur) > 0 {
			paras = append(paras, strings.Join(cur, " "))
			cur = cur[:0:0]
		}
	}
	for _, line := range strings.Split(s, "\n") {
		// Fields режет по unicode.IsSpace: сюда входят таб, \r и U+00A0.
		words := strings.Fields(line)
		if len(words) == 0 {
			// Строка из одних пробелов — такой же разделитель, как пустая.
			flush()
			continue
		}
		cur = append(cur, words...)
	}
	flush()
	return strings.Join(paras, "\n\n")
}
