package main

func chkGoodOrder() Order {
	return Order{ID: "o1", Shipping: &Address{"Казань", "420000"}, Items: []Item{{"A", 1}, {"B", 2}}}
}

func TestValidateOrderOK(t *testing.T) {
	if err := ValidateOrder(chkGoodOrder()); err != nil {
		t.Fatalf("корректный заказ: ожидали nil, получили %#v", err)
	}
}

func TestPrefix(t *testing.T) {
	city := &FieldError{"City", ErrRequired}
	err := Prefix("Shipping", errors.Join(city, &FieldError{"Zip", ErrFormat}))
	if got := Paths(err); !slices.Equal(got, []string{"Shipping.City", "Shipping.Zip"}) {
		t.Fatalf("Paths(Prefix) = %v, ожидали [Shipping.City Shipping.Zip]", got)
	}
	if city.Path != "City" {
		t.Fatalf("Prefix изменил исходную ошибку: Path = %q", city.Path)
	}
	if !errors.Is(err, ErrRequired) || !errors.Is(err, ErrFormat) {
		t.Fatal("после Prefix errors.Is должен находить ErrRequired и ErrFormat")
	}
	if Prefix("X", nil) != nil {
		t.Fatal("Prefix(nil) должен вернуть nil")
	}
	nested := Prefix("Items[0]", Prefix("Box", &FieldError{"W", ErrRange}))
	if got := Paths(nested); !slices.Equal(got, []string{"Items[0].Box.W"}) {
		t.Fatalf("вложенный Prefix: %v, ожидали [Items[0].Box.W]", got)
	}
}

func TestValidateOrderPaths(t *testing.T) {
	o := Order{
		Shipping: &Address{"", "12a456"},
		Items:    []Item{{"A", 0}, {"B", 1}, {"A", 1}, {"", 101}, {"B", 5}},
	}
	err := ValidateOrder(o)
	want := []string{"ID", "Shipping.City", "Shipping.Zip", "Items[0].Qty", "Items[2].SKU", "Items[3].SKU", "Items[3].Qty", "Items[4].SKU"}
	if got := Paths(err); !slices.Equal(got, want) {
		t.Fatalf("пути ошибок\n%v\nожидали\n%v", got, want)
	}
	for _, s := range []string{"Items[2].SKU: дубликат: как Items[0]", "Items[4].SKU: дубликат: как Items[1]", "Shipping.Zip: неверный формат"} {
		if !strings.Contains(err.Error(), s) {
			t.Fatalf("в тексте ошибки нет %q:\n%v", s, err)
		}
	}
	if !errors.Is(err, ErrDuplicate) || !errors.Is(err, ErrRange) {
		t.Fatal("errors.Is должен находить ErrDuplicate и ErrRange")
	}
	var fe *FieldError
	if !errors.As(err, &fe) || fe.Path != "ID" {
		t.Fatalf("errors.As должен найти первый *FieldError (ID), нашёл %v", fe)
	}
}

func TestValidateOrderMissing(t *testing.T) {
	err := ValidateOrder(Order{ID: "o2"})
	if got := Paths(err); !slices.Equal(got, []string{"Shipping", "Items"}) {
		t.Fatalf("без адреса и позиций: пути %v, ожидали [Shipping Items]", got)
	}
}
