package main

type chkAddr struct {
	City string `json:"city"`
	Zip  int    `json:"zip"`
}

type chkPatch struct {
	Nick  Field[string]  `json:"nick"`
	Age   Field[int]     `json:"age"`
	Addr  Field[chkAddr] `json:"addr"`
	Score Field[float64] `json:"score"`
}

func chkDecode(t *testing.T, s string) chkPatch {
	t.Helper()
	var p chkPatch
	if err := json.Unmarshal([]byte(s), &p); err != nil {
		t.Fatalf("json.Unmarshal(%s): %v", s, err)
	}
	return p
}

func TestFieldThreeStates(t *testing.T) {
	p := chkDecode(t, `{"nick": null, "age": 0, "addr": {"city": "Тверь", "zip": 170000}}`)
	if p.Score.Present {
		t.Fatalf("score нет в JSON, а Present = true")
	}
	if !p.Nick.Present || !p.Nick.Null {
		t.Fatalf("nick: null → ожидали Present и Null, получили %+v", p.Nick)
	}
	if v, ok := p.Age.Get(); !ok || v != 0 || !p.Age.Present || p.Age.Null {
		t.Fatalf("age: 0 — это значение, а не отсутствие: %+v, Get = (%v, %v)", p.Age, v, ok)
	}
	if v, ok := p.Addr.Get(); !ok || v != (chkAddr{"Тверь", 170000}) {
		t.Fatalf("addr: вложенный объект = (%+v, %v)", v, ok)
	}
	if _, ok := p.Nick.Get(); ok {
		t.Fatalf("Get у null должен вернуть false")
	}
	if _, ok := p.Score.Get(); ok {
		t.Fatalf("Get у отсутствующего поля должен вернуть false")
	}
}

func TestFieldBadType(t *testing.T) {
	var p chkPatch
	if err := json.Unmarshal([]byte(`{"age": "двадцать"}`), &p); err == nil {
		t.Fatalf("строка в int-поле должна дать ошибку, а её нет: %+v", p.Age)
	}
}

func TestFieldApply(t *testing.T) {
	type row struct {
		Nick *string
		Age  *int
	}
	old := "neo"
	oldAge := 30
	r := row{Nick: &old, Age: &oldAge}

	p := chkDecode(t, `{"age": 31}`)
	p.Nick.Apply(&r.Nick)
	p.Age.Apply(&r.Age)
	if r.Nick != &old || r.Age == nil || *r.Age != 31 || oldAge != 30 {
		t.Fatalf("{age: 31}: nick должен остаться прежним указателем, age стать 31, старое значение не трогать: nick=%v age=%v old=%d", r.Nick, r.Age, oldAge)
	}
	p.Age.Value = 99
	if *r.Age != 31 {
		t.Fatalf("после Apply колонка делит память с полем запроса: %d", *r.Age)
	}

	p = chkDecode(t, `{"nick": null}`)
	p.Nick.Apply(&r.Nick)
	p.Age.Apply(&r.Age)
	if r.Nick != nil || r.Age == nil || *r.Age != 31 {
		t.Fatalf("{nick: null}: nick должен стать nil, age остаться 31: nick=%v age=%v", r.Nick, r.Age)
	}
}
