package main

type chkUser struct {
	ID, Email string
	Age       int
}

func (u chkUser) Key() string { return u.ID }
func (u chkUser) Validate() error {
	if !strings.Contains(u.Email, "@") {
		return chkBadEmail
	}
	return nil
}

var chkBadEmail = errors.New("bad email")

// chkDoc — сущность по указателю, с версией.
type chkDoc struct {
	Name string
	Ver  int
}

func (d *chkDoc) Key() string  { return d.Name }
func (d *chkDoc) Version() int { return d.Ver }

type chkTag string

func (t chkTag) Key() string { return string(t) }

func TestRepoBasic(t *testing.T) {
	r := NewRepo[chkUser]()
	if err := r.Save(chkUser{"u2", "b@x", 30}); err != nil {
		t.Fatalf("Save = %v", err)
	}
	r.Save(chkUser{"u1", "a@x", 20})
	r.Save(chkUser{"u3", "c@x", 40})
	u, err := r.Get("u1")
	if err != nil || u.Age != 20 {
		t.Fatalf("Get(u1) = %+v, %v", u, err)
	}
	u, err = r.Get("zz")
	if !errors.Is(err, ErrNotFound) || !strings.Contains(err.Error(), `"zz"`) || u != (chkUser{}) {
		t.Fatalf("Get(zz) = %+v, %v; ожидали нулевое значение и ErrNotFound с ключом в тексте", u, err)
	}
	var keys []string
	for _, u := range r.List(nil) {
		keys = append(keys, u.ID)
	}
	if strings.Join(keys, ",") != "u1,u2,u3" {
		t.Fatalf("List(nil) = %v, ожидали u1,u2,u3 по возрастанию ключа", keys)
	}
	old := r.List(func(u chkUser) bool { return u.Age >= 30 })
	if len(old) != 2 || old[0].ID != "u2" {
		t.Fatalf("List(age>=30) = %+v", old)
	}
	if got := r.List(func(chkUser) bool { return false }); got == nil || len(got) != 0 {
		t.Fatalf("пустой результат List = %#v, ожидали пустой не-nil срез", got)
	}
	if err := r.Delete("u2"); err != nil {
		t.Fatalf("Delete(u2) = %v", err)
	}
	if err := r.Delete("u2"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("повторный Delete = %v, ожидали ErrNotFound", err)
	}
}

func TestRepoValidate(t *testing.T) {
	r := NewRepo[chkUser]()
	err := r.Save(chkUser{"u1", "no-at", 1})
	if !errors.Is(err, chkBadEmail) {
		t.Fatalf("Save невалидного = %v; T реализует Validator — ожидали ошибку Validate", err)
	}
	if _, err := r.Get("u1"); !errors.Is(err, ErrNotFound) {
		t.Fatal("невалидная сущность не должна сохраняться")
	}
	tags := NewRepo[chkTag]() // без Validator и Versioned — просто сохраняется
	if err := tags.Save("go"); err != nil {
		t.Fatalf("Save для типа без опциональных интерфейсов = %v", err)
	}
}

func TestRepoVersions(t *testing.T) {
	r := NewRepo[*chkDoc]()
	if err := r.Save(&chkDoc{"a", 2}); !errors.Is(err, ErrConflict) {
		t.Fatalf("новая запись с версией 2: %v, ожидали ErrConflict", err)
	}
	if err := r.Save(&chkDoc{"a", 1}); err != nil {
		t.Fatalf("новая запись с версией 1: %v", err)
	}
	if err := r.Save(&chkDoc{"a", 2}); err != nil {
		t.Fatalf("обновление 1→2: %v", err)
	}
	err := r.Save(&chkDoc{"a", 2})
	if !errors.Is(err, ErrConflict) || !strings.Contains(err.Error(), `"a"`) {
		t.Fatalf("устаревшая версия 2 поверх 2: %v; ожидали ErrConflict с ключом", err)
	}
	if err := r.Save(&chkDoc{"a", 4}); !errors.Is(err, ErrConflict) {
		t.Fatalf("скачок версии 2→4: %v, ожидали ErrConflict", err)
	}
	if d, _ := r.Get("a"); d == nil || d.Ver != 2 {
		t.Fatalf("после конфликтов хранится %+v, ожидали версию 2", d)
	}
}

func TestRepoConcurrent(t *testing.T) {
	r := NewRepo[chkTag]()
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range 50 {
				r.Save(chkTag(fmt.Sprint(i*100 + j)))
				r.List(nil)
			}
		}()
	}
	wg.Wait()
	if n := len(r.List(nil)); n != 400 {
		t.Fatalf("после параллельных Save в репозитории %d записей, ожидали 400", n)
	}
}
