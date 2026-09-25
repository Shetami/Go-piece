package main

func TestGroupAnagramsBasic(t *testing.T) {
	got := GroupAnagrams([]string{"eat", "tea", "tan", "ate", "nat", "bat"})
	want := [][]string{{"eat", "tea", "ate"}, {"tan", "nat"}, {"bat"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("получили %q, ожидали %q", got, want)
	}
}

func TestGroupAnagramsCyrillic(t *testing.T) {
	got := GroupAnagrams([]string{"Пила", "липа", "кот", "ток", "пал"})
	want := [][]string{{"Пила", "липа"}, {"кот", "ток"}, {"пал"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("получили %q, ожидали %q", got, want)
	}
}

func TestGroupAnagramsCountMatters(t *testing.T) {
	got := GroupAnagrams([]string{"aab", "abb"})
	if len(got) != 2 {
		t.Fatalf("aab и abb — не анаграммы: получили %q", got)
	}
}

func TestGroupAnagramsEmpty(t *testing.T) {
	if got := GroupAnagrams(nil); len(got) != 0 {
		t.Fatalf("пустой вход: %q", got)
	}
}

func TestGroupAnagramsStableOrder(t *testing.T) {
	in := []string{"b", "a", "c", "a"}
	want := [][]string{{"b"}, {"a", "a"}, {"c"}}
	for range 10 {
		if got := GroupAnagrams(in); !reflect.DeepEqual(got, want) {
			t.Fatalf("порядок групп плавает: %q, ожидали %q", got, want)
		}
	}
}
