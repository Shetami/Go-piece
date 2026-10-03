// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/word-count, MIT, © Exercism). Руками не править.
package main

func TestWordCount(t *testing.T) {
	cases := []struct {
		name     string
		sentence string
		want     map[string]int
	}{
		{"count one word", "word", map[string]int{"word": 1}},
		{"count one of each word", "one of each", map[string]int{"each": 1, "of": 1, "one": 1}},
		{"multiple occurrences of a word", "one fish two fish red fish blue fish", map[string]int{"blue": 1, "fish": 4, "one": 1, "red": 1, "two": 1}},
		{"handles cramped lists", "one,two,three", map[string]int{"one": 1, "three": 1, "two": 1}},
		{"handles expanded lists", "one,\ntwo,\nthree", map[string]int{"one": 1, "three": 1, "two": 1}},
		{"ignore punctuation", "car: carpet as java: javascript!!&@$%^&", map[string]int{"as": 1, "car": 1, "carpet": 1, "java": 1, "javascript": 1}},
		{"include numbers", "testing, 1, 2 testing", map[string]int{"1": 1, "2": 1, "testing": 2}},
		{"normalize case", "go Go GO Stop stop", map[string]int{"go": 3, "stop": 2}},
		{"with apostrophes", "'First: don't laugh. Then: don't cry. You're getting it.'", map[string]int{"cry": 1, "don't": 2, "first": 1, "getting": 1, "it": 1, "laugh": 1, "then": 1, "you're": 1}},
		{"with quotations", "Joe can't tell between 'large' and large.", map[string]int{"and": 1, "between": 1, "can't": 1, "joe": 1, "large": 2, "tell": 1}},
		{"substrings from the beginning", "Joe can't tell between app, apple and a.", map[string]int{"a": 1, "and": 1, "app": 1, "apple": 1, "between": 1, "can't": 1, "joe": 1, "tell": 1}},
		{"multiple spaces not detected as a word", " multiple   whitespaces", map[string]int{"multiple": 1, "whitespaces": 1}},
		{"alternating word separators not detected as a word", ",\n,one,\n ,two \n 'three'", map[string]int{"one": 1, "three": 1, "two": 1}},
		{"quotation for word with apostrophe", "can, can't, 'can't'", map[string]int{"can": 1, "can't": 2}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := WordCount(c.sentence)
			if !sameResult(got, c.want) {
				t.Fatalf("WordCount(%q) = %v, ожидали %v", c.sentence, got, c.want)
			}
		})
	}
}

// sameResult — reflect.DeepEqual, который не отличает nil от пустого слайса
// или мапы: для ответа это одно и то же.
func sameResult(got, want any) bool {
	g, w := reflect.ValueOf(got), reflect.ValueOf(want)
	if g.Kind() == w.Kind() && (g.Kind() == reflect.Slice || g.Kind() == reflect.Map) && g.Len() == 0 && w.Len() == 0 {
		return true
	}
	return reflect.DeepEqual(got, want)
}
