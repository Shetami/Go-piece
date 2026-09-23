package main

func TestTopWordsBasic(t *testing.T) {
	got := TopWords("go go go rust rust zig", 2)
	if want := []string{"go", "rust"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("TopWords = %v, ожидали %v", got, want)
	}
}

func TestTopWordsCaseAndPunctuation(t *testing.T) {
	got := TopWords("Go, go! GO? Rust... rust; zig", 3)
	if want := []string{"go", "rust", "zig"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("TopWords с регистром и пунктуацией = %v, ожидали %v", got, want)
	}
}

func TestTopWordsTiesAlphabetical(t *testing.T) {
	got := TopWords("яблоко груша банан груша яблоко банан", 3)
	if want := []string{"банан", "груша", "яблоко"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("при равной частоте — по алфавиту: %v, ожидали %v", got, want)
	}
}

func TestTopWordsKTooBig(t *testing.T) {
	got := TopWords("один два два", 10)
	if want := []string{"два", "один"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("k больше числа слов: %v, ожидали %v", got, want)
	}
	if got := TopWords("", 3); len(got) != 0 {
		t.Fatalf("пустой текст: %v", got)
	}
}
