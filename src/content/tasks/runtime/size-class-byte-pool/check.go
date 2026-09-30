package main

func chkClass(n int) int {
	c := MinClass
	for c < n {
		c *= 2
	}
	return c
}

func TestBytePoolFresh(t *testing.T) {
	var p BytePool
	for _, n := range []int{0, 1, 64, 65, 100, 128, 1000, 4096, 65535, 65536} {
		b := p.Get(n)
		if len(b) != n || cap(b) != chkClass(n) {
			t.Fatalf("Get(%d): len=%d cap=%d, ожидали len=%d cap=%d", n, len(b), cap(b), n, chkClass(n))
		}
	}
	if b := p.Get(70000); len(b) != 70000 {
		t.Fatalf("Get(70000): len=%d", len(b))
	}
}

func TestBytePoolOddCap(t *testing.T) {
	var p BytePool
	for i := 0; i < 50; i++ {
		p.Put(make([]byte, 10, 100))
		p.Put(make([]byte, 0, 200))
	}
	for i := 0; i < 100; i++ {
		for _, n := range []int{64, 100, 120, 128, 200, 256} {
			var b []byte
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("Get(%d) после Put слайсов с cap 100 и 200 паникует: %v", n, r)
					}
				}()
				b = p.Get(n)
			}()
			if len(b) != n || cap(b) != chkClass(n) {
				t.Fatalf("Get(%d) после Put слайсов с cap 100 и 200: len=%d cap=%d, ожидали cap=%d", n, len(b), cap(b), chkClass(n))
			}
		}
	}
}

func TestBytePoolReuse(t *testing.T) {
	var p BytePool
	reused := 0
	for i := 0; i < 50; i++ {
		b := p.Get(1000)
		b[0] = 42
		p.Put(b)
		if c := p.Get(1000); &c[0] == &b[0] {
			reused++
		}
	}
	if reused == 0 {
		t.Fatalf("за 50 циклов Put→Get слайс ни разу не вернулся из пула")
	}
}

func TestBytePoolRejects(t *testing.T) {
	var p BytePool
	for i := 0; i < 20; i++ {
		p.Put(make([]byte, 1<<20))
		p.Put(make([]byte, 10))
		p.Put(nil)
		if b := p.Get(MaxClass); cap(b) != MaxClass {
			t.Fatalf("Get(%d) вернул cap=%d — мегабайтный слайс не должен попадать в пул", MaxClass, cap(b))
		}
		if b := p.Get(5); cap(b) != MinClass {
			t.Fatalf("Get(5) вернул cap=%d, ожидали %d", cap(b), MinClass)
		}
	}
}

func TestBytePoolConcurrent(t *testing.T) {
	var p BytePool
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 300; i++ {
				n := (g*977 + i*131) % 5000
				b := p.Get(n)
				if len(b) != n || cap(b) != chkClass(n) {
					t.Errorf("Get(%d): len=%d cap=%d", n, len(b), cap(b))
					return
				}
				p.Put(b)
			}
		}(g)
	}
	wg.Wait()
}
