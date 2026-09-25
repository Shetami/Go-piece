package main

func TestWrapBasic(t *testing.T) {
	got := Wrap("the quick brown fox jumps over the lazy dog", 10)
	want := []string{"the quick", "brown fox", "jumps over", "the lazy", "dog"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("получили %q, ожидали %q", got, want)
	}
}

func TestWrapCyrillicWidth(t *testing.T) {
	got := Wrap("съешь же ещё этих мягких булок", 12)
	want := []string{"съешь же ещё", "этих мягких", "булок"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ширина считается в символах: получили %q, ожидали %q", got, want)
	}
}

func TestWrapWhitespace(t *testing.T) {
	got := Wrap("  a\tb\n\nc   ", 80)
	if !reflect.DeepEqual(got, []string{"a b c"}) {
		t.Fatalf("лишние пробелы: получили %q, ожидали [\"a b c\"]", got)
	}
}

func TestWrapLongWord(t *testing.T) {
	got := Wrap("a supercalifragilistic b", 5)
	want := []string{"a", "supercalifragilistic", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("длинное слово: получили %q, ожидали %q", got, want)
	}
}

func TestWrapEmpty(t *testing.T) {
	if got := Wrap("   ", 10); len(got) != 0 {
		t.Fatalf("пустой текст: %q", got)
	}
}
