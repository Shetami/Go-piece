package main

type chkEntity struct {
	name string
	res  *chkRes
}

type chkRes struct{ buf [256]byte }

func TestPoolBasic(t *testing.T) {
	var p Pool[string]
	if _, ok := p.Get(Handle{}); ok {
		t.Fatalf("Get(Handle{}) на пустом пуле вернул ok")
	}
	a := p.Insert("a")
	b := p.Insert("b")
	if _, ok := p.Get(Handle{}); ok {
		t.Fatalf("нулевой Handle{} не должен ссылаться на живой объект")
	}
	if v, ok := p.Get(a); !ok || *v != "a" {
		t.Fatalf("Get(a) неверно")
	}
	v, _ := p.Get(b)
	*v = "B"
	if v, _ := p.Get(b); *v != "B" {
		t.Fatalf("Get возвращает не указатель на хранимый объект")
	}
	if p.Len() != 2 {
		t.Fatalf("Len = %d, ожидали 2", p.Len())
	}
}

func TestPoolStaleHandle(t *testing.T) {
	var p Pool[string]
	old := p.Insert("старый")
	if !p.Remove(old) {
		t.Fatalf("Remove живого объекта вернул false")
	}
	if p.Remove(old) {
		t.Fatalf("повторный Remove вернул true")
	}
	fresh := p.Insert("новый")
	if p.Slots() != 1 {
		t.Fatalf("Slots = %d, ожидали 1 — освободившийся слот должен переиспользоваться", p.Slots())
	}
	if v, ok := p.Get(old); ok {
		t.Fatalf("старый Handle после переиспользования слота вернул %q", *v)
	}
	if p.Remove(old) {
		t.Fatalf("Remove по старому Handle удалил чужой объект")
	}
	if v, ok := p.Get(fresh); !ok || *v != "новый" {
		t.Fatalf("новый объект пропал")
	}
}

func TestPoolDoubleRemove(t *testing.T) {
	var p Pool[int]
	h := p.Insert(1)
	p.Remove(h)
	p.Remove(h)
	x := p.Insert(10)
	y := p.Insert(20)
	if x == y {
		t.Fatalf("после двойного Remove два Insert вернули одинаковый Handle")
	}
	vx, _ := p.Get(x)
	vy, _ := p.Get(y)
	if vx == nil || vy == nil || *vx != 10 || *vy != 20 {
		t.Fatalf("после двойного Remove два объекта делят один слот")
	}
	if p.Len() != 2 {
		t.Fatalf("Len = %d, ожидали 2", p.Len())
	}
}

func TestPoolAll(t *testing.T) {
	var p Pool[int]
	var hs []Handle
	for i := 0; i < 6; i++ {
		hs = append(hs, p.Insert(i))
	}
	p.Remove(hs[1])
	p.Remove(hs[4])
	var got []int
	for h, v := range p.All() {
		if w, ok := p.Get(h); !ok || w != v {
			t.Fatalf("All выдал Handle, не соответствующий объекту")
		}
		got = append(got, *v)
	}
	if !slices.Equal(got, []int{0, 2, 3, 5}) {
		t.Fatalf("All = %v, ожидали [0 2 3 5]", got)
	}
	n := 0
	for range p.All() {
		n++
		break
	}
	if n != 1 {
		t.Fatalf("All не остановился после break")
	}
}

func TestPoolReleases(t *testing.T) {
	var p Pool[chkEntity]
	var freed atomic.Int32
	var hs []Handle
	for i := 0; i < 20; i++ {
		r := &chkRes{}
		runtime.SetFinalizer(r, func(*chkRes) { freed.Add(1) })
		hs = append(hs, p.Insert(chkEntity{name: "e", res: r}))
	}
	for _, h := range hs {
		p.Remove(h)
	}
	for i := 0; i < 50 && freed.Load() < 20; i++ {
		runtime.GC()
		time.Sleep(time.Millisecond)
	}
	if got := freed.Load(); got < 20 {
		t.Fatalf("после Remove и GC собрано %d из 20 ресурсов — пул удерживает удалённые объекты", got)
	}
	runtime.KeepAlive(&p)
}
