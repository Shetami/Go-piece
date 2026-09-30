package main

func chkSlug(t *testing.T, in string, maxLen int, want string) {
	t.Helper()
	if got := Slug(in, maxLen); got != want {
		t.Fatalf("Slug(%q, %d) = %q, ожидали %q", in, maxLen, got, want)
	}
}

func TestSlugTranslit(t *testing.T) {
	chkSlug(t, "Привет, мир", 0, "privet-mir")
	chkSlug(t, "Щука и ЁЖ", 0, "shchuka-i-yozh")
	chkSlug(t, "ЩИ да каша", 0, "shchi-da-kasha")
	chkSlug(t, "Подъезд, объявление и соль", 0, "podezd-obyavlenie-i-sol")
	chkSlug(t, "Go 1.23: что нового?", 0, "go-1-23-chto-novogo")
}

func TestSlugSeparators(t *testing.T) {
	chkSlug(t, "  --Привет!!!   мир...  ", 0, "privet-mir")
	chkSlug(t, "snake_case и kebab-case", 0, "snake-case-i-kebab-case")
	chkSlug(t, "Привет,мир", 0, "privet-mir")
	chkSlug(t, "!!!", 0, "")
	chkSlug(t, "", 0, "")
	chkSlug(t, "Ь", 0, "")
}

func TestSlugForeignLetters(t *testing.T) {
	chkSlug(t, "Café Müller", 0, "caf-mller")
	chkSlug(t, "日本 Go", 0, "go")
}

func TestSlugMaxLen(t *testing.T) {
	chkSlug(t, "Щука и ёж: две истории", 14, "shchuka-i-yozh")
	chkSlug(t, "Щука и ёж: две истории", 15, "shchuka-i-yozh")
	chkSlug(t, "Щука и ёж: две истории", 13, "shchuka-i")
	chkSlug(t, "Щука и ёж", 5, "shchu")
	chkSlug(t, "Щука и ёж", 100, "shchuka-i-yozh")
	chkSlug(t, "Щука и ёж", -1, "shchuka-i-yozh")
	for _, n := range []int{3, 8, 10, 20} {
		got := Slug("Длинный заголовок статьи про строки в Go", n)
		if len(got) > n || strings.HasSuffix(got, "-") || strings.HasPrefix(got, "-") {
			t.Fatalf("Slug(…, %d) = %q: длиннее лимита или с дефисом на краю", n, got)
		}
	}
}
