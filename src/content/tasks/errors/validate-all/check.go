package main

func fields(err error) []string {
	var out []string
	var walk func(error)
	walk = func(e error) {
		if fe, ok := e.(*FieldError); ok {
			out = append(out, fe.Field)
		}
		if j, ok := e.(interface{ Unwrap() []error }); ok {
			for _, x := range j.Unwrap() {
				walk(x)
			}
		} else if x := errors.Unwrap(e); x != nil {
			walk(x)
		}
	}
	walk(err)
	return out
}

func TestValidateOK(t *testing.T) {
	if err := Validate(User{"аня", "a@b.ru", 30}); err != nil {
		t.Fatalf("корректный пользователь: %v", err)
	}
}

func TestValidateAll(t *testing.T) {
	err := Validate(User{"", "нет-почты", 200})
	got := fields(err)
	if !reflect.DeepEqual(got, []string{"Name", "Email", "Age"}) {
		t.Fatalf("ожидали ошибки по полям [Name Email Age] в этом порядке, получили %v (%v)", got, err)
	}
	for _, s := range []string{"Name: обязательно", "Email: нет @", "Age: вне диапазона"} {
		if !strings.Contains(err.Error(), s) {
			t.Fatalf("в тексте %q нет %q", err, s)
		}
	}
}

func TestValidateAs(t *testing.T) {
	err := Validate(User{"боря", "b@c.ru", -1})
	var fe *FieldError
	if !errors.As(err, &fe) || fe.Field != "Age" {
		t.Fatalf("errors.As должен найти *FieldError для Age, получили %v", err)
	}
}

func TestValidateNilIsNil(t *testing.T) {
	err := Validate(User{"вика", "v@x.ru", 0})
	if err != nil {
		t.Fatalf("возраст 0 допустим, ожидали nil, получили %#v", err)
	}
}
