package main

func chkIDs(n int) []int {
	ids := make([]int, n)
	for i := range ids {
		ids[i] = i + 1
	}
	return ids
}

func TestProcessAllOK(t *testing.T) {
	var cur, peak atomic.Int32
	rep, err := ProcessAll(context.Background(), chkIDs(20), 3, func(ctx context.Context, id int) error {
		c := cur.Add(1)
		for {
			p := peak.Load()
			if c <= p || peak.CompareAndSwap(p, c) {
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
		cur.Add(-1)
		return nil
	})
	if err != nil || !slices.Equal(rep.Done, chkIDs(20)) {
		t.Fatalf("Done = %v, err = %v; ожидали все 20 id по возрастанию и nil", rep.Done, err)
	}
	if peak.Load() > 3 {
		t.Fatalf("одновременно работало %d handle, лимит 3", peak.Load())
	}
	if rep.Failed == nil || len(rep.Skipped) != 0 {
		t.Fatalf("Failed должна быть пустой (не nil) мапой, Skipped пустым: %v, %v", rep.Failed, rep.Skipped)
	}
}

func TestProcessAllFailed(t *testing.T) {
	boom := errors.New("битые данные")
	downstream := fmt.Errorf("вызов склада: %w", context.Canceled)
	rep, err := ProcessAll(context.Background(), chkIDs(6), 2, func(ctx context.Context, id int) error {
		switch id {
		case 2:
			return boom
		case 5:
			return downstream // чужая отмена при живом ctx — это провал
		}
		return nil
	})
	if err != nil {
		t.Fatalf("ctx не отменяли, а ProcessAll вернула %v", err)
	}
	if !slices.Equal(rep.Done, []int{1, 3, 4, 6}) || len(rep.Failed) != 2 ||
		!errors.Is(rep.Failed[2], boom) || !errors.Is(rep.Failed[5], context.Canceled) || len(rep.Skipped) != 0 {
		t.Fatalf("отчёт %+v; ожидали Done [1 3 4 6], Failed {2, 5}, Skipped пусто", rep)
	}
}

func TestProcessAllCancel(t *testing.T) {
	stop := errors.New("сервис останавливается")
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	rep, err := ProcessAll(ctx, chkIDs(10), 1, func(ctx context.Context, id int) error {
		if id == 3 {
			cancel(stop)
			return ctx.Err()
		}
		return nil
	})
	if !slices.Equal(rep.Done, []int{1, 2}) || !slices.Equal(rep.Skipped, chkIDs(10)[2:]) || len(rep.Failed) != 0 {
		t.Fatalf("отчёт %+v; ожидали Done [1 2], Skipped [3..10] — ни один id не должен потеряться", rep)
	}
	if !errors.Is(err, stop) {
		t.Fatalf("err = %v, ожидали причину отмены (context.Cause)", err)
	}
}

func TestProcessAllPreCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var calls atomic.Int32
	rep, err := ProcessAll(ctx, chkIDs(5), 4, func(context.Context, int) error { calls.Add(1); return nil })
	if calls.Load() != 0 {
		t.Fatalf("ctx отменён заранее, а handle вызвана %d раз", calls.Load())
	}
	if !slices.Equal(rep.Skipped, chkIDs(5)) || !errors.Is(err, context.Canceled) {
		t.Fatalf("Skipped = %v, err = %v; ожидали все id и Canceled", rep.Skipped, err)
	}
}

func TestProcessAllWaits(t *testing.T) {
	var finished atomic.Int32
	ProcessAll(context.Background(), chkIDs(8), 4, func(context.Context, int) error {
		time.Sleep(10 * time.Millisecond)
		finished.Add(1)
		return nil
	})
	if finished.Load() != 8 {
		t.Fatalf("ProcessAll вернулась, когда закончились только %d из 8 handle", finished.Load())
	}
}
