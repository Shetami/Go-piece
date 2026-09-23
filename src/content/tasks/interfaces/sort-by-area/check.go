package main

var _ sort.Interface = ByArea(nil)

func TestSortByArea(t *testing.T) {
	shapes := []Shape{Rect{3, 3}, Circle{1}, Rect{1, 2}, Circle{2}}
	sort.Sort(ByArea(shapes))
	want := []Shape{Rect{1, 2}, Circle{1}, Rect{3, 3}, Circle{2}}
	if !reflect.DeepEqual(shapes, want) {
		t.Fatalf("после сортировки %v, ожидали %v", shapes, want)
	}
}

func TestSortByAreaStable(t *testing.T) {
	shapes := []Shape{Rect{2, 1}, Rect{1, 1}, Rect{1, 2}}
	sort.Stable(ByArea(shapes))
	want := []Shape{Rect{1, 1}, Rect{2, 1}, Rect{1, 2}}
	if !reflect.DeepEqual(shapes, want) {
		t.Fatalf("sort.Stable: %v, ожидали %v", shapes, want)
	}
}

func TestSortByAreaLen(t *testing.T) {
	if n := ByArea([]Shape{Circle{1}, Rect{1, 1}}).Len(); n != 2 {
		t.Fatalf("Len = %d, ожидали 2", n)
	}
}
