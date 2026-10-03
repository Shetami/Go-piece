// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/anagram, MIT, © Exercism). Руками не править.
package main

func TestAnagrams(t *testing.T) {
	cases := []struct {
		name       string
		subject    string
		candidates []string
		want       []string
	}{
		{"no matches", "diaper", []string{"hello", "world", "zombies", "pants"}, []string{}},
		{"detects two anagrams", "solemn", []string{"lemons", "cherry", "melons"}, []string{"lemons", "melons"}},
		{"does not detect anagram subsets", "good", []string{"dog", "goody"}, []string{}},
		{"detects anagram", "listen", []string{"enlists", "google", "inlets", "banana"}, []string{"inlets"}},
		{"detects three anagrams", "allergy", []string{"gallery", "ballerina", "regally", "clergy", "largely", "leading"}, []string{"gallery", "regally", "largely"}},
		{"detects multiple anagrams with different case", "nose", []string{"Eons", "ONES"}, []string{"Eons", "ONES"}},
		{"does not detect non-anagrams with identical checksum", "mass", []string{"last"}, []string{}},
		{"detects anagrams case-insensitively", "Orchestra", []string{"cashregister", "Carthorse", "radishes"}, []string{"Carthorse"}},
		{"detects anagrams using case-insensitive subject", "Orchestra", []string{"cashregister", "carthorse", "radishes"}, []string{"carthorse"}},
		{"detects anagrams using case-insensitive possible matches", "orchestra", []string{"cashregister", "Carthorse", "radishes"}, []string{"Carthorse"}},
		{"does not detect an anagram if the original word is repeated", "go", []string{"goGoGO"}, []string{}},
		{"anagrams must use all letters exactly once", "tapper", []string{"patter"}, []string{}},
		{"words are not anagrams of themselves", "BANANA", []string{"BANANA"}, []string{}},
		{"words are not anagrams of themselves even if letter case is partially different", "BANANA", []string{"Banana"}, []string{}},
		{"words are not anagrams of themselves even if letter case is completely different", "BANANA", []string{"banana"}, []string{}},
		{"words other than themselves can be anagrams", "LISTEN", []string{"LISTEN", "Silent"}, []string{"Silent"}},
		{"handles case of greek letters", "ΑΒΓ", []string{"ΒΓΑ", "ΒΓΔ", "γβα", "αβγ"}, []string{"ΒΓΑ", "γβα"}},
		{"different characters may have the same bytes", "a⬂", []string{"€a"}, []string{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Anagrams(c.subject, c.candidates)
			if !sameResult(got, c.want) {
				t.Fatalf("Anagrams(%q, %q) = %q, ожидали %q", c.subject, c.candidates, got, c.want)
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
