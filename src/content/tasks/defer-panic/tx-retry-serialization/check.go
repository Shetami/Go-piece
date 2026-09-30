package main

type chkDB struct {
	log        []string
	commitErrs []error // ошибка Commit для каждой попытки по порядку
	rbErr      error
}

type chkTx struct{ db *chkDB }

func (d *chkDB) Begin(ctx context.Context) (Tx, error) {
	d.log = append(d.log, "begin")
	return &chkTx{d}, nil
}

func (t *chkTx) Commit() error {
	t.db.log = append(t.db.log, "commit")
	if len(t.db.commitErrs) > 0 {
		e := t.db.commitErrs[0]
		t.db.commitErrs = t.db.commitErrs[1:]
		return e
	}
	return nil
}

func (t *chkTx) Rollback() error { t.db.log = append(t.db.log, "rollback"); return t.db.rbErr }

func chkLog(t *testing.T, db *chkDB, want ...string) {
	t.Helper()
	if !reflect.DeepEqual(db.log, want) {
		t.Fatalf("вызовы %v,\nожидали %v", db.log, want)
	}
}

func TestTxRetryOnFnConflict(t *testing.T) {
	db := &chkDB{}
	calls := 0
	err := RunInTx(context.Background(), db, 3, func(ctx context.Context, tx Tx) error {
		calls++
		if calls == 1 {
			return fmt.Errorf("update balance: %w", ErrSerialization)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("вторая попытка успешна, а RunInTx вернул %v", err)
	}
	chkLog(t, db, "begin", "rollback", "begin", "commit")
}

func TestTxRetryOnCommitConflict(t *testing.T) {
	db := &chkDB{commitErrs: []error{fmt.Errorf("pq: %w", ErrSerialization)}}
	if err := RunInTx(context.Background(), db, 3, func(context.Context, Tx) error { return nil }); err != nil {
		t.Fatalf("RunInTx = %v, ожидали nil со второй попытки", err)
	}
	chkLog(t, db, "begin", "commit", "begin", "commit")
}

func TestTxGivesUp(t *testing.T) {
	db := &chkDB{}
	err := RunInTx(context.Background(), db, 2, func(context.Context, Tx) error { return ErrSerialization })
	if !errors.Is(err, ErrSerialization) {
		t.Fatalf("после исчерпания попыток ожидали ErrSerialization, получили %v", err)
	}
	chkLog(t, db, "begin", "rollback", "begin", "rollback")
}

func TestTxNoRetryAndRollbackError(t *testing.T) {
	rbErr := errors.New("соединение оборвано")
	db := &chkDB{rbErr: rbErr}
	bad := errors.New("недостаточно средств")
	err := RunInTx(context.Background(), db, 5, func(context.Context, Tx) error { return bad })
	if !errors.Is(err, bad) || !errors.Is(err, rbErr) {
		t.Fatalf("err = %v, ожидали и ошибку fn, и ошибку Rollback", err)
	}
	chkLog(t, db, "begin", "rollback")
}

func TestTxPanicNoRetry(t *testing.T) {
	db := &chkDB{}
	defer func() {
		if r := recover(); r != "nil pointer в бизнес-логике" {
			t.Fatalf("паника должна пролететь с тем же значением, recover() = %v", r)
		}
		chkLog(t, db, "begin", "rollback")
	}()
	RunInTx(context.Background(), db, 3, func(context.Context, Tx) error { panic("nil pointer в бизнес-логике") })
	t.Fatal("RunInTx проглотил панику")
}

func TestTxContextCancel(t *testing.T) {
	db := &chkDB{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := RunInTx(ctx, db, 5, func(context.Context, Tx) error {
		cancel()
		return ErrSerialization
	})
	if !errors.Is(err, context.Canceled) || !errors.Is(err, ErrSerialization) {
		t.Fatalf("err = %v, ожидали context.Canceled вместе с ошибкой попытки", err)
	}
	chkLog(t, db, "begin", "rollback")
}
