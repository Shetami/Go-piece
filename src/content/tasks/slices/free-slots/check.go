package main

func chkFree(t *testing.T, busy []Interval, day Interval, minLen int, want []Interval) {
	t.Helper()
	orig := slices.Clone(busy)
	got := FreeSlots(busy, day, minLen)
	if len(got) == 0 && len(want) == 0 {
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FreeSlots(%v, %v, %d) = %v, ожидали %v", orig, day, minLen, got, want)
	}
	if !reflect.DeepEqual(busy, orig) {
		t.Fatalf("FreeSlots изменил busy: %v, был %v", busy, orig)
	}
}

var chkDay = Interval{540, 1080} // 9:00–18:00

func TestFreeSlotsBasic(t *testing.T) {
	chkFree(t, []Interval{{720, 780}, {600, 660}}, chkDay, 30,
		[]Interval{{540, 600}, {660, 720}, {780, 1080}})
	chkFree(t, nil, chkDay, 30, []Interval{chkDay})
}

func TestFreeSlotsOverlapNested(t *testing.T) {
	// {600,900} поглощает {650,700}; {880,950} перекрывает его хвост.
	chkFree(t, []Interval{{650, 700}, {600, 900}, {880, 950}}, chkDay, 1,
		[]Interval{{540, 600}, {950, 1080}})
}

func TestFreeSlotsTouchingAndMinLen(t *testing.T) {
	// Касающиеся встречи не оставляют между собой пустой щели.
	chkFree(t, []Interval{{600, 660}, {660, 720}}, chkDay, 1,
		[]Interval{{540, 600}, {720, 1080}})
	// Ровно minLen подходит, на минуту меньше — нет.
	chkFree(t, []Interval{{570, 600}, {629, 1080}}, chkDay, 30,
		[]Interval{{540, 570}})
}

func TestFreeSlotsOutsideDay(t *testing.T) {
	chkFree(t, []Interval{{480, 600}, {1000, 1200}, {0, 60}, {1300, 1400}}, chkDay, 1,
		[]Interval{{600, 1000}})
	chkFree(t, []Interval{{0, 2000}}, chkDay, 1, nil)
}

func TestFreeSlotsEmptyIntervals(t *testing.T) {
	// {700,650} и {800,800} пустые — они не должны ничего разрезать.
	chkFree(t, []Interval{{700, 650}, {800, 800}}, chkDay, 1, []Interval{chkDay})
	chkFree(t, []Interval{{600, 700}, {650, 640}}, chkDay, 1,
		[]Interval{{540, 600}, {700, 1080}})
}
