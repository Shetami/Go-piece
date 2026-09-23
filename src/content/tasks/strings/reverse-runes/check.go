package main

func TestReverseASCII(t *testing.T) {
	if got := Reverse("golang"); got != "gnalog" {
		t.Fatalf(`Reverse("golang") = %q, ожидали "gnalog"`, got)
	}
}

func TestReverseEmptyAndSingle(t *testing.T) {
	if got := Reverse(""); got != "" {
		t.Fatalf(`Reverse("") = %q, ожидали ""`, got)
	}
	if got := Reverse("я"); got != "я" {
		t.Fatalf(`Reverse("я") = %q, ожидали "я"`, got)
	}
}

func TestReverseCyrillic(t *testing.T) {
	// Здесь разворот по байтам и развалится: каждая буква — два байта.
	if got := Reverse("Привет"); got != "тевирП" {
		t.Fatalf(`Reverse("Привет") = %q, ожидали "тевирП" (похоже, строка развёрнута по байтам)`, got)
	}
}

func TestReverseMixedWidth(t *testing.T) {
	if got := Reverse("aя日"); got != "日яa" {
		t.Fatalf(`Reverse("aя日") = %q, ожидали "日яa"`, got)
	}
	if got := Reverse(Reverse("Go — это круто")); got != "Go — это круто" {
		t.Fatalf("двойной разворот не вернул исходную строку: %q", got)
	}
}
