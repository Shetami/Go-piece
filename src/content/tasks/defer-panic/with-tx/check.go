package main

type checkTx struct {
	calls     []string
	commitErr error
}

func (t *checkTx) Commit() error   { t.calls = append(t.calls, "commit"); return t.commitErr }
func (t *checkTx) Rollback() error { t.calls = append(t.calls, "rollback"); return nil }

func TestWithTxCommit(t *testing.T) {
	tx := &checkTx{}
	if err := WithTx(tx, func() error { return nil }); err != nil {
		t.Fatalf("успешная fn: ожидали nil, получили %v", err)
	}
	if !reflect.DeepEqual(tx.calls, []string{"commit"}) {
		t.Fatalf("вызовы %v, ожидали [commit]", tx.calls)
	}
}

func TestWithTxRollbackOnError(t *testing.T) {
	tx := &checkTx{}
	boom := errors.New("не хватило денег")
	err := WithTx(tx, func() error { return boom })
	if !errors.Is(err, boom) {
		t.Fatalf("ожидали ошибку fn, получили %v", err)
	}
	if !reflect.DeepEqual(tx.calls, []string{"rollback"}) {
		t.Fatalf("вызовы %v, ожидали [rollback]", tx.calls)
	}
}

func TestWithTxCommitError(t *testing.T) {
	commitErr := errors.New("сериализация не удалась")
	tx := &checkTx{commitErr: commitErr}
	if err := WithTx(tx, func() error { return nil }); !errors.Is(err, commitErr) {
		t.Fatalf("ошибка Commit должна вернуться, получили %v", err)
	}
}

func TestWithTxPanic(t *testing.T) {
	tx := &checkTx{}
	defer func() {
		r := recover()
		if r != "упали" {
			t.Fatalf("паника должна пройти дальше с тем же значением, получили %v", r)
		}
		if !reflect.DeepEqual(tx.calls, []string{"rollback"}) {
			t.Fatalf("при панике вызовы %v, ожидали [rollback]", tx.calls)
		}
	}()
	WithTx(tx, func() error { panic("упали") })
	t.Fatal("WithTx проглотил панику")
}
