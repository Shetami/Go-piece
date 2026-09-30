package main

import (
	"slices"
	"strings"
	"unicode"
)

// Highlight выделяет в text все вхождения терминов terms тегами
// <mark>…</mark> — как сниппет в результатах поиска.
//
//   - поиск без учёта регистра, посимвольно: руна текста и руна термина
//     равны, если равны их unicode.ToLower («Го» находит «гО», «ГО»);
//   - вхождения ищутся с перекрытием: термин "аа" в "ааа" покрывает весь "ааа";
//   - пересекающиеся и соприкасающиеся вхождения (разных и одного
//     термина) сливаются в один <mark>: "ab" и "cd" в "abcd" → "<mark>abcd</mark>";
//   - регистр текста в результате сохраняется;
//   - пустые термины игнорируются;
//   - символы & < > " в тексте экранируются (&amp; &lt; &gt; &quot;),
//     в том числе внутри выделения; искать термины нужно в исходном тексте.
func Highlight(text string, terms []string) string {
	src := []rune(text)
	// Нижний регистр — по руне: число рун не меняется, индексы совпадают
	// с src. strings.ToLower всей строки может поменять длину в байтах.
	low := make([]rune, len(src))
	for i, r := range src {
		low[i] = unicode.ToLower(r)
	}
	type span struct{ from, to int } // [from, to) в рунах
	var spans []span
	for _, term := range terms {
		tr := []rune(term)
		if len(tr) == 0 {
			continue
		}
		for i := range tr {
			tr[i] = unicode.ToLower(tr[i])
		}
		for i := 0; i+len(tr) <= len(low); i++ { // с перекрытием: шаг 1
			if slices.Equal(low[i:i+len(tr)], tr) {
				spans = append(spans, span{i, i + len(tr)})
			}
		}
	}
	// Сливаем пересекающиеся и соприкасающиеся интервалы.
	slices.SortFunc(spans, func(a, b span) int { return a.from - b.from })
	var merged []span
	for _, s := range spans {
		if n := len(merged); n > 0 && s.from <= merged[n-1].to {
			merged[n-1].to = max(merged[n-1].to, s.to)
			continue
		}
		merged = append(merged, s)
	}
	var b strings.Builder
	pos := 0
	for _, s := range merged {
		writeEscaped(&b, src[pos:s.from])
		b.WriteString("<mark>")
		writeEscaped(&b, src[s.from:s.to])
		b.WriteString("</mark>")
		pos = s.to
	}
	writeEscaped(&b, src[pos:])
	return b.String()
}

func writeEscaped(b *strings.Builder, rs []rune) {
	for _, r := range rs {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		default:
			b.WriteRune(r)
		}
	}
}
