package main

func TestTryMapAllOK(t *testing.T) {
	got, err := TryMap([]string{"1", "20", "-3"}, strconv.Atoi)
	if err != nil {
		t.Fatalf("все строки — числа, а err = %#v (ожидали настоящий nil)", err)
	}
	if !reflect.DeepEqual(got, []int{1, 20, -3}) {
		t.Fatalf("TryMap = %v, ожидали [1 20 -3]", got)
	}
}

func TestTryMapCollectsAll(t *testing.T) {
	calls := 0
	got, err := TryMap([]string{"10", "x", "30", "4.5", "50"}, func(s string) (int, error) {
		calls++
		return strconv.Atoi(s)
	})
	if calls != 5 {
		t.Fatalf("f вызвана %d раз, ожидали 5 — после ошибки не останавливаемся", calls)
	}
	if !reflect.DeepEqual(got, []int{10, 30, 50}) {
		t.Fatalf("значения = %v, ожидали только успешные по порядку [10 30 50]", got)
	}
	if err == nil {
		t.Fatalf("две строки не числа, а err = nil")
	}
	if !errors.Is(err, strconv.ErrSyntax) {
		t.Fatalf("errors.Is(err, strconv.ErrSyntax) = false — исходная ошибка потеряна: %v", err)
	}
	var ie *ItemError
	if !errors.As(err, &ie) || ie.Index != 1 {
		t.Fatalf("errors.As до *ItemError: %v, первая ошибка должна быть с Index 1", ie)
	}
	multi, ok := err.(interface{ Unwrap() []error })
	if !ok {
		t.Fatalf("ошибки нужно объединить errors.Join, а получили %T", err)
	}
	var idx []int
	for _, e := range multi.Unwrap() {
		if errors.As(e, &ie) {
			idx = append(idx, ie.Index)
		}
	}
	if !reflect.DeepEqual(idx, []int{1, 3}) {
		t.Fatalf("индексы ошибок = %v, ожидали [1 3]", idx)
	}
	want := "элемент 1: strconv.Atoi: parsing \"x\": invalid syntax\nэлемент 3: strconv.Atoi: parsing \"4.5\": invalid syntax"
	if err.Error() != want {
		t.Fatalf("текст ошибки:\n%s\nожидали:\n%s", err.Error(), want)
	}
}

func TestCollectIndexes(t *testing.T) {
	boom := errors.New("таймаут")
	rs := []Result[float64]{{Val: 1.5}, {Err: boom}, {Val: 0}}
	vals, err := Collect(rs)
	var ie *ItemError
	if !reflect.DeepEqual(vals, []float64{1.5, 0}) || !errors.As(err, &ie) || ie.Index != 1 || !errors.Is(err, boom) {
		t.Fatalf("Collect = %v, %v — ожидали [1.5 0] и ошибку элемента 1", vals, err)
	}
	if vals, err := Collect([]Result[int]{}); err != nil || len(vals) != 0 {
		t.Fatalf("Collect(пусто) = %v, %#v — ожидали пусто и nil", vals, err)
	}
}
