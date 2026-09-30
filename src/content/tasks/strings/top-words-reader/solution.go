package main

import (
	"bufio"
	"cmp"
	"fmt"
	"io"
	"slices"
	"strings"
	"unicode"
)

// WordCount — слово и сколько раз оно встретилось.
type WordCount struct {
	Word  string
	Count int
}

// TopWords читает текст из r и возвращает k самых частых слов.
//
//   - слово — последовательность букв и цифр; апостроф (' или ’) и дефис
//     между двумя буквами/цифрами — часть слова ("don't", "кто-то"),
//     в любом другом месте — разделитель ("-то", "rock'n'roll'" → "rock'n'roll");
//   - регистр не важен, 'ё' считается как 'е'; слова возвращаются в нижнем
//     регистре и с 'е' вместо 'ё';
//   - порядок: по убыванию Count, при равенстве — по возрастанию Word
//     (strings.Compare);
//   - k <= 0 или пустой текст — nil;
//   - r может отдавать данные любыми кусками: руна или слово может
//     разорваться между вызовами Read; текст может быть больше памяти
//     под одну строку, поэтому читать потоково;
//   - ошибка чтения возвращается обёрнутой, результат при этом nil.
func TopWords(r io.Reader, k int) ([]WordCount, error) {
	br := bufio.NewReader(r) // ReadRune склеивает руны, разрезанные между Read
	counts := map[string]int{}
	var word strings.Builder
	var pending rune // апостроф или дефис после слова: решим, когда увидим следующий символ
	flush := func() {
		if word.Len() > 0 {
			counts[word.String()]++
			word.Reset()
		}
		pending = 0
	}
	for {
		c, _, err := br.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("чтение текста: %w", err)
		}
		switch {
		case unicode.IsLetter(c) || unicode.IsDigit(c):
			c = unicode.ToLower(c)
			if c == 'ё' {
				c = 'е'
			}
			if pending != 0 {
				word.WriteRune(pending) // знак оказался между буквами — часть слова
				pending = 0
			}
			word.WriteRune(c)
		case (c == '\'' || c == '’' || c == '-') && word.Len() > 0 && pending == 0:
			pending = c
		default:
			flush()
		}
	}
	flush()
	if k <= 0 || len(counts) == 0 {
		return nil, nil
	}
	out := make([]WordCount, 0, len(counts))
	for w, n := range counts {
		out = append(out, WordCount{w, n})
	}
	slices.SortFunc(out, func(a, b WordCount) int {
		return cmp.Or(cmp.Compare(b.Count, a.Count), strings.Compare(a.Word, b.Word))
	})
	return out[:min(k, len(out))], nil
}
