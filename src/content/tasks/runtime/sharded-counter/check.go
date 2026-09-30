package main

func TestShardSize(t *testing.T) {
	if s := unsafe.Sizeof(shard{}); s == 0 || s%cacheLine != 0 {
		t.Fatalf("unsafe.Sizeof(shard{}) = %d, ожидали кратное %d — соседние шарды делят строку кэша", s, cacheLine)
	}
}

func TestCounterShards(t *testing.T) {
	prev := runtime.GOMAXPROCS(3)
	defer runtime.GOMAXPROCS(prev)
	c := NewCounter()
	if c.Shards() != 3 {
		t.Fatalf("при GOMAXPROCS=3 шардов %d, ожидали 3", c.Shards())
	}
	runtime.GOMAXPROCS(5)
	if c.Shards() != 3 {
		t.Fatalf("число шардов должно фиксироваться при создании")
	}
	if got := NewCounter().Shards(); got != 5 {
		t.Fatalf("NewCounter при GOMAXPROCS=5: %d шардов", got)
	}
	if runtime.GOMAXPROCS(0) != 5 {
		t.Fatalf("NewCounter изменил GOMAXPROCS")
	}
}

func TestCounterSum(t *testing.T) {
	c := NewCounter()
	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				if g%4 == 0 {
					c.Add(-1)
				} else {
					c.Add(2)
				}
			}
		}(g)
	}
	wg.Wait()
	if got, want := c.Value(), int64(12*2000-4*1000); got != want {
		t.Fatalf("Value = %d, ожидали %d", got, want)
	}
}

func TestCounterAddNoAlloc(t *testing.T) {
	c := NewCounter()
	if a := testing.AllocsPerRun(1000, func() { c.Add(1) }); a != 0 {
		t.Fatalf("Add выделяет память %.2f раз на вызов, ожидали 0", a)
	}
	if c.Value() != 1001 {
		t.Fatalf("Value = %d, ожидали 1001 (1000 замеров + прогрев)", c.Value())
	}
}
