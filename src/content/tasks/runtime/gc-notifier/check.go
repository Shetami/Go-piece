package main

func chkWaitCount(c *atomic.Int32, want int32) bool {
	for i := 0; i < 200; i++ {
		if c.Load() >= want {
			return true
		}
		runtime.Gosched()
		time.Sleep(time.Millisecond)
	}
	return false
}

func TestOnGCEveryCycle(t *testing.T) {
	var calls atomic.Int32
	stop := OnGC(func() { calls.Add(1) })
	defer stop()
	for i := int32(1); i <= 5; i++ {
		runtime.GC()
		if !chkWaitCount(&calls, i) {
			t.Fatalf("после %d-й сборки мусора f вызвана %d раз — уведомление должно приходить после каждого цикла", i, calls.Load())
		}
	}
}

func TestOnGCStop(t *testing.T) {
	var calls atomic.Int32
	stop := OnGC(func() { calls.Add(1) })
	runtime.GC()
	if !chkWaitCount(&calls, 1) {
		t.Fatalf("после runtime.GC() f не вызвана")
	}
	stop()
	for i := 0; i < 3; i++ {
		runtime.GC()
		time.Sleep(2 * time.Millisecond)
	}
	if n := calls.Load(); n > 2 {
		t.Fatalf("после stop f вызвана ещё %d раз", n-1)
	}
}

func TestOnGCIndependent(t *testing.T) {
	var a, b atomic.Int32
	stopA := OnGC(func() { a.Add(1) })
	stopB := OnGC(func() { b.Add(1) })
	defer stopB()
	runtime.GC()
	if !chkWaitCount(&a, 1) || !chkWaitCount(&b, 1) {
		t.Fatalf("два уведомителя: a=%d, b=%d после GC, ожидали оба >= 1", a.Load(), b.Load())
	}
	stopA()
	before := a.Load()
	for i := int32(2); i <= 4; i++ {
		runtime.GC()
		if !chkWaitCount(&b, i) {
			t.Fatalf("остановка одного уведомителя сломала другой: b=%d после %d сборок", b.Load(), i)
		}
	}
	if a.Load() > before+1 {
		t.Fatalf("остановленный уведомитель продолжает вызываться: %d → %d", before, a.Load())
	}
}
