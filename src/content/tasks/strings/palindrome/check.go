package main

func TestPalindromeLatin(t *testing.T) {
	for _, s := range []string{"racecar", "Was it a car or a cat I saw?", "No 'x' in Nixon"} {
		if !IsPalindrome(s) {
			t.Fatalf("%q — палиндром", s)
		}
	}
	if IsPalindrome("golang") {
		t.Fatal(`"golang" — не палиндром`)
	}
}

func TestPalindromeCyrillic(t *testing.T) {
	if !IsPalindrome("А роза упала на лапу Азора") {
		t.Fatal("«А роза упала на лапу Азора» — палиндром")
	}
	if IsPalindrome("Привет") {
		t.Fatal("«Привет» — не палиндром")
	}
}

func TestPalindromeEdge(t *testing.T) {
	for _, s := range []string{"", "я", "!!!", "12321"} {
		if !IsPalindrome(s) {
			t.Fatalf("%q — палиндром (пустое и из одного символа — тоже)", s)
		}
	}
	if IsPalindrome("1231") {
		t.Fatal(`"1231" — не палиндром`)
	}
}
