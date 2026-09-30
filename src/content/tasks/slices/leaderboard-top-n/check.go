package main

// chkRef — заведомо правильная медленная модель.
type chkRef struct {
	n  int
	es []Entry
}

func (r *chkRef) submit(name string, score int) {
	for i, e := range r.es {
		if e.Name == name {
			if score <= e.Score {
				return
			}
			r.es = append(r.es[:i:i], r.es[i+1:]...)
			break
		}
	}
	r.es = append(r.es, Entry{name, score})
	sort.SliceStable(r.es, func(i, j int) bool { return r.es[i].Score > r.es[j].Score })
	if len(r.es) > r.n {
		r.es = r.es[:r.n]
	}
}

func TestBoardBasic(t *testing.T) {
	b := NewBoard(3)
	b.Submit("ann", 50)
	b.Submit("bob", 70)
	b.Submit("cat", 50) // равный счёт позже — ниже ann
	want := []Entry{{"bob", 70}, {"ann", 50}, {"cat", 50}}
	if got := b.Top(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Top = %v, ожидали %v", got, want)
	}
	b.Submit("dan", 50) // равен последнему в полной таблице — не входит
	if got := b.Top(); !reflect.DeepEqual(got, want) {
		t.Fatalf("после dan=50: %v, ожидали %v", got, want)
	}
	b.Submit("ann", 40) // хуже своего — игнор
	b.Submit("cat", 80) // лучше — переезжает, без дубля
	want = []Entry{{"cat", 80}, {"bob", 70}, {"ann", 50}}
	if got := b.Top(); !reflect.DeepEqual(got, want) {
		t.Fatalf("после cat=80: %v, ожидали %v", got, want)
	}
}

func TestBoardEvictedReturns(t *testing.T) {
	b := NewBoard(2)
	b.Submit("a", 10)
	b.Submit("b", 20)
	b.Submit("c", 30) // a вылетает
	b.Submit("a", 25) // a забыт и возвращается с новым счётом
	want := []Entry{{"c", 30}, {"a", 25}}
	if got := b.Top(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Top = %v, ожидали %v — вылетевший игрок должен забываться", got, want)
	}
	b.Submit("b", 5) // b вылетел раньше; 5 в полную таблицу не входит
	if got := b.Top(); !reflect.DeepEqual(got, want) {
		t.Fatalf("после b=5: %v, ожидали %v", got, want)
	}
	top := b.Top()
	top[0].Score = 1
	if b.Top()[0].Score != 30 {
		t.Fatalf("Top должен возвращать копию")
	}
}

func TestBoardAgainstModel(t *testing.T) {
	b := NewBoard(5)
	ref := &chkRef{n: 5}
	seed := uint32(1)
	for step := range 3000 {
		seed = seed*1664525 + 1013904223
		name := string(rune('a' + (seed>>20)%9))
		seed = seed*1664525 + 1013904223
		score := int(seed>>24) % 40
		b.Submit(name, score)
		ref.submit(name, score)
		if got := b.Top(); !reflect.DeepEqual(got, ref.es) && !(len(got) == 0 && len(ref.es) == 0) {
			t.Fatalf("шаг %d, Submit(%s, %d): Top = %v, ожидали %v", step, name, score, got, ref.es)
		}
	}
}
