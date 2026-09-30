package main

func chkCatalog(ids ...string) []Item {
	out := make([]Item, len(ids))
	for i, id := range ids {
		out[i] = Item{ID: id, Title: "t" + id}
	}
	return out
}

func chkPageIDs(p []Item) []string {
	ids := []string{}
	for _, it := range p {
		ids = append(ids, it.ID)
	}
	return ids
}

func TestPageWalk(t *testing.T) {
	items := chkCatalog("a", "a1", "b", "c", "d", "e", "f")
	var all []string
	after := ""
	for step := 0; ; step++ {
		p, next := Page(items, after, 3)
		all = append(all, chkPageIDs(p)...)
		if next == "" {
			break
		}
		if step > 10 {
			t.Fatalf("обход по курсору не заканчивается: next = %q", next)
		}
		after = next
	}
	if want := chkPageIDs(items); !reflect.DeepEqual(all, want) {
		t.Fatalf("обход по страницам дал %v, ожидали %v", all, want)
	}
}

func TestPageNextAtEnd(t *testing.T) {
	items := chkCatalog("a", "b", "c", "d")
	p, next := Page(items, "b", 2)
	if got := chkPageIDs(p); !reflect.DeepEqual(got, []string{"c", "d"}) || next != "" {
		t.Fatalf("Page(after=b, 2) = %v, next=%q; ожидали [c d] и пустой next — дальше ничего нет", got, next)
	}
	if p, next := Page(items, "d", 5); len(p) != 0 || next != "" {
		t.Fatalf("после последнего: %v, next=%q", chkPageIDs(p), next)
	}
	if p, next := Page(items, "", 0); len(p) != 0 || next != "" {
		t.Fatalf("limit=0: %v, next=%q", chkPageIDs(p), next)
	}
}

func TestPageDeletedCursor(t *testing.T) {
	items := chkCatalog("a", "c", "e", "g")
	p, next := Page(items, "d", 2) // "d" удалили между запросами
	if got := chkPageIDs(p); !reflect.DeepEqual(got, []string{"e", "g"}) || next != "" {
		t.Fatalf("курсор d (удалён): %v, next=%q; ожидали [e g]", got, next)
	}
	p, next = Page(items, "b", 1)
	if got := chkPageIDs(p); !reflect.DeepEqual(got, []string{"c"}) || next != "c" {
		t.Fatalf("курсор b (удалён): %v, next=%q; ожидали [c], next=c", got, next)
	}
}

func TestPageAppendSafe(t *testing.T) {
	items := chkCatalog("a", "b", "c", "d")
	p, _ := Page(items, "", 2)
	_ = append(p, Item{ID: "zzz"})
	if items[2].ID != "c" {
		t.Fatalf("append к странице затёр items[2]: %v", chkPageIDs(items))
	}
}

func TestPageLogarithmic(t *testing.T) {
	n := 200000
	items := make([]Item, n)
	for i := range items {
		items[i] = Item{ID: fmt.Sprintf("id%07d", i)}
	}
	start := time.Now()
	for i := 0; i < 50000; i++ {
		p, _ := Page(items, items[(i*37)%n].ID, 1)
		if len(p) > 0 && p[0].ID != items[(i*37)%n+1].ID {
			t.Fatalf("после %s получили %s", items[(i*37)%n].ID, p[0].ID)
		}
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("50000 страниц из %d элементов заняли %v — начало страницы ищут перебором", n, d)
	}
}
