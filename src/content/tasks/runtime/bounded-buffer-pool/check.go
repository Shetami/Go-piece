package main

func TestBufferPoolGetEmpty(t *testing.T) {
	p := NewBufferPool(1024)
	b := p.Get()
	if b == nil {
		t.Fatalf("Get на пустом пуле вернул nil, ожидали новый буфер")
	}
	for i := 0; i < 50; i++ {
		b.WriteString("старые данные")
		p.Put(b)
		b = p.Get()
		if b == nil || b.Len() != 0 {
			t.Fatalf("Get вернул буфер с данными %q, ожидали пустой", b.String())
		}
	}
}

func TestBufferPoolReuses(t *testing.T) {
	p := NewBufferPool(1024)
	reused := 0
	for i := 0; i < 50; i++ {
		b := p.Get()
		b.WriteString("x")
		p.Put(b)
		if p.Get() == b {
			reused++
		}
	}
	if reused == 0 {
		t.Fatalf("за 50 циклов Put→Get буфер ни разу не вернулся из пула — пул не переиспользует память")
	}
}

func TestBufferPoolDropsBig(t *testing.T) {
	p := NewBufferPool(1024)
	for i := 0; i < 20; i++ {
		big := new(bytes.Buffer)
		big.Grow(1 << 20)
		p.Put(big)
		if b := p.Get(); b.Cap() > 1024 {
			t.Fatalf("Get вернул буфер с Cap=%d при maxCap=1024 — большие буферы нельзя класть в пул", b.Cap())
		}
	}
}

func TestBufferPoolNilAndConcurrent(t *testing.T) {
	p := NewBufferPool(4096)
	p.Put(nil)
	if b := p.Get(); b == nil {
		t.Fatalf("после Put(nil) Get вернул nil")
	}
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				b := p.Get()
				if b.Len() != 0 {
					t.Errorf("конкурентный Get вернул непустой буфер")
					return
				}
				b.WriteString("данные")
				p.Put(b)
			}
		}()
	}
	wg.Wait()
}
