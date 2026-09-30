package main

var chkErrFunds = errors.New("недостаточно средств")

func TestRepoErrors(t *testing.T) {
	r := NewRepo()
	if _, err := r.Get("u1"); !errors.Is(err, ErrNotFound) || !strings.Contains(fmt.Sprint(err), "u1") {
		t.Fatalf("Get несуществующего: ожидали ErrNotFound с ID в тексте, получили %v", err)
	}
	if err := r.Create(User{ID: "u1", Name: "Аня"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := r.Create(User{ID: "u1"}); !errors.Is(err, ErrExists) {
		t.Fatalf("повторный Create: ожидали ErrExists, получили %v", err)
	}
	if _, err := r.Update(User{ID: "u9", Version: 1}); !errors.Is(err, ErrNotFound) || errors.Is(err, ErrConflict) {
		t.Fatalf("Update несуществующего: ожидали ErrNotFound (не конфликт), получили %v", err)
	}
}

func TestRepoVersion(t *testing.T) {
	r := NewRepo()
	r.Create(User{ID: "u1", Name: "Аня"})
	u, _ := r.Get("u1")
	if u.Version != 1 {
		t.Fatalf("после Create версия %d, ожидали 1", u.Version)
	}
	u.Name = "Анна"
	saved, err := r.Update(u)
	if err != nil || saved.Version != 2 {
		t.Fatalf("Update = %+v, %v; ожидали версию 2", saved, err)
	}
	_, err = r.Update(u) // устаревшая версия 1
	var ve *VersionError
	if !errors.Is(err, ErrConflict) || !errors.As(err, &ve) || ve.Want != 1 || ve.Have != 2 {
		t.Fatalf("устаревший Update: ожидали *VersionError{Want:1 Have:2}, который Is(ErrConflict), получили %v", err)
	}
	if err.Error() != "user u1: версия 1, в базе 2" {
		t.Fatalf("текст %q, ожидали %q", err.Error(), "user u1: версия 1, в базе 2")
	}
}

func TestModifyBusinessError(t *testing.T) {
	r := NewRepo()
	r.Create(User{ID: "u1", Balance: 10})
	calls := 0
	err := Modify(r, "u1", func(u *User) error {
		calls++
		if u.Balance < 100 {
			return fmt.Errorf("списать 100: %w", chkErrFunds)
		}
		return nil
	})
	if !errors.Is(err, chkErrFunds) || calls != 1 {
		t.Fatalf("ошибка fn возвращается сразу: err=%v вызовов=%d", err, calls)
	}
	if u, _ := r.Get("u1"); u.Version != 1 {
		t.Fatalf("при ошибке fn Update не вызывается, а версия стала %d", u.Version)
	}
	if err := Modify(r, "nope", func(*User) error { return nil }); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Modify несуществующего: ожидали ErrNotFound, получили %v", err)
	}
}

func TestModifyConcurrent(t *testing.T) {
	r := NewRepo()
	r.Create(User{ID: "u1"})
	var wg sync.WaitGroup
	var failed atomic.Int64
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := Modify(r, "u1", func(u *User) error {
				time.Sleep(time.Millisecond)
				u.Balance++
				return nil
			})
			if err != nil {
				failed.Add(1)
			}
		}()
	}
	wg.Wait()
	u, _ := r.Get("u1")
	if failed.Load() != 0 || u.Balance != 20 || u.Version != 21 {
		t.Fatalf("20 параллельных пополнений: неудач %d, баланс %d, версия %d; ожидали 0, 20, 21 (потерянные обновления?)", failed.Load(), u.Balance, u.Version)
	}
}
