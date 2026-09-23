package main

func TestFilterInts(t *testing.T) {
	got := Filter([]int{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 })
	if want := []int{2, 4, 6}; !reflect.DeepEqual(got, want) {
		t.Fatalf("чётные из 1..6 = %v, ожидали %v", got, want)
	}
}

func TestFilterKeepsOrder(t *testing.T) {
	got := Filter([]string{"го", "а", "б", "ланг"}, func(s string) bool { return len(s) > 2 })
	if want := []string{"го", "ланг"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Filter по строкам = %v, ожидали %v (порядок должен сохраняться)", got, want)
	}
}

func TestFilterDoesNotTouchInput(t *testing.T) {
	src := []int{5, 1, 4, 2}
	before := append([]int(nil), src...)

	Filter(src, func(n int) bool { return n > 2 })

	if !reflect.DeepEqual(src, before) {
		t.Fatalf("исходный слайс изменился: был %v, стал %v", before, src)
	}
}

func TestFilterEdgeCases(t *testing.T) {
	if got := Filter([]int{1, 2, 3}, func(int) bool { return false }); len(got) != 0 {
		t.Fatalf("когда ничего не подошло, ожидали пустой результат, получили %v", got)
	}
	if got := Filter([]int(nil), func(int) bool { return true }); len(got) != 0 {
		t.Fatalf("из nil-слайса ожидали пустой результат, получили %v", got)
	}

	type user struct {
		Name string
		Age  int
	}
	users := []user{{"Аня", 30}, {"Боря", 17}}
	got := Filter(users, func(u user) bool { return u.Age >= 18 })
	if want := []user{{"Аня", 30}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Filter по структурам = %v, ожидали %v", got, want)
	}
}
