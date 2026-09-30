package main

var chkNoDigits = ValidatorFunc(func(s string) error {
	if strings.ContainsAny(s, "0123456789") {
		return errors.New("digits not allowed")
	}
	return nil
})

func TestValidatorFunc(t *testing.T) {
	var v Validator = chkNoDigits
	if v.Validate("abc") != nil || v.Validate("a1") == nil {
		t.Fatal("ValidatorFunc.Validate должен вызывать саму функцию")
	}
}

func TestMinLenRunes(t *testing.T) {
	if err := MinLen(6).Validate("привет"); err != nil {
		t.Fatalf("MinLen(6) на «привет» (6 букв, 12 байт) = %v; длина считается в символах", err)
	}
	err := MinLen(5).Validate("кот")
	if err == nil || err.Error() != "length 3 is less than 5" {
		t.Fatalf("MinLen(5) на «кот» = %v, ожидали \"length 3 is less than 5\"", err)
	}
}

func TestNotBlank(t *testing.T) {
	for _, s := range []string{"", "   ", "\t\n"} {
		if NotBlank().Validate(s) == nil {
			t.Fatalf("NotBlank пропустил %q", s)
		}
	}
	if NotBlank().Validate(" x ") != nil {
		t.Fatal("NotBlank отверг \" x \"")
	}
}

func TestAll(t *testing.T) {
	v := All(NotBlank(), nil, MinLen(3), chkNoDigits)
	if err := v.Validate("abc"); err != nil {
		t.Fatalf("All на \"abc\" = %v, ожидали nil (nil-валидатор пропускается)", err)
	}
	err := v.Validate("7")
	if err == nil || err.Error() != "length 1 is less than 3\ndigits not allowed" {
		t.Fatalf("All на \"7\" = %q; ожидали обе ошибки по порядку через errors.Join", err)
	}
	if err := All().Validate(""); err != nil {
		t.Fatalf("All() без валидаторов = %v, ожидали nil", err)
	}
	nested := All(All(NotBlank()), MinLen(2))
	if err := nested.Validate(" "); err == nil || !strings.Contains(err.Error(), "blank") {
		t.Fatalf("вложенный All = %v", err)
	}
}
