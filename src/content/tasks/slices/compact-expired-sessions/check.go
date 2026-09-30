package main

var chkNow = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

func chkSess(id string, d time.Duration) *Session {
	return &Session{ID: id, Expires: chkNow.Add(d)}
}

func chkIDs(s []*Session) []string {
	ids := []string{}
	for _, x := range s {
		ids = append(ids, x.ID)
	}
	return ids
}

func TestDropExpiredOrder(t *testing.T) {
	in := []*Session{chkSess("a", time.Hour), chkSess("b", -time.Hour), nil, chkSess("c", 0), chkSess("d", time.Second)}
	got := chkIDs(DropExpired(in, chkNow))
	if want := []string{"a", "d"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("остались %v, ожидали %v (Expires == now — уже истекла, nil выбрасываем)", got, want)
	}
}

func TestDropExpiredInPlaceAndTail(t *testing.T) {
	in := []*Session{chkSess("a", -1), chkSess("b", time.Hour), chkSess("c", -1), chkSess("d", time.Hour), chkSess("e", -1)}
	got := DropExpired(in, chkNow)
	if len(got) != 2 || &got[0] != &in[0] {
		t.Fatalf("результат должен лежать в начале того же массива: len=%d", len(got))
	}
	for i := len(got); i < len(in); i++ {
		if in[i] != nil {
			t.Fatalf("in[%d] = %s: хвост массива держит выброшенную сессию, сборщик мусора её не освободит", i, in[i].ID)
		}
	}
}

func TestDropExpiredAllGone(t *testing.T) {
	in := []*Session{chkSess("a", -1), nil, chkSess("b", -time.Minute)}
	if got := DropExpired(in, chkNow); len(got) != 0 {
		t.Fatalf("все истекли, а осталось %v", chkIDs(got))
	}
	if in[0] != nil || in[2] != nil {
		t.Fatalf("все истекли, но массив всё ещё держит указатели")
	}
	if got := DropExpired(nil, chkNow); len(got) != 0 {
		t.Fatalf("DropExpired(nil) = %v", got)
	}
}

func TestDropExpiredNoAlloc(t *testing.T) {
	base := make([]*Session, 100)
	for i := range base {
		base[i] = chkSess(fmt.Sprint(i), time.Duration(i%3-1)*time.Hour)
	}
	buf := make([]*Session, len(base))
	allocs := testing.AllocsPerRun(50, func() {
		copy(buf, base)
		DropExpired(buf, chkNow)
	})
	if allocs > 0 {
		t.Fatalf("DropExpired выделяет память: %.0f аллокаций на вызов, ожидали 0", allocs)
	}
}
