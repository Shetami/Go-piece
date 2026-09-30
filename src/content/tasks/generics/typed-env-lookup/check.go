package main

type chkLevel int

func (l *chkLevel) UnmarshalText(b []byte) error {
	switch string(b) {
	case "debug":
		*l = 0
	case "info":
		*l = 1
	case "error":
		*l = 2
	default:
		return fmt.Errorf("неизвестный уровень %q", b)
	}
	return nil
}

type chkPort int

var chkEnv = map[string]string{
	"PORT":    "8080",
	"RATIO":   "0.75",
	"DEBUG":   "true",
	"TIMEOUT": "1m30s",
	"BIG":     "9000000000",
	"NAME":    "",
	"LEVEL":   "error",
	"ADDR":    "10.0.0.7",
	"START":   "2024-03-01T10:00:00Z",
	"BAD_INT": "80a",
	"EMPTY":   "",
	"YES":     "yes",
}

func TestLookupBasic(t *testing.T) {
	if v, err := Lookup(chkEnv, "PORT", 80); v != 8080 || err != nil {
		t.Fatalf("PORT = (%v, %v)", v, err)
	}
	if v, err := Lookup(chkEnv, "RATIO", 1.0); v != 0.75 || err != nil {
		t.Fatalf("RATIO = (%v, %v)", v, err)
	}
	if v, err := Lookup(chkEnv, "DEBUG", false); !v || err != nil {
		t.Fatalf("DEBUG = (%v, %v)", v, err)
	}
	if v, err := Lookup(chkEnv, "BIG", int64(0)); v != 9000000000 || err != nil {
		t.Fatalf("BIG = (%v, %v)", v, err)
	}
	if v, err := Lookup(chkEnv, "NAME", "anon"); v != "" || err != nil {
		t.Fatalf("NAME задан пустым — это значение, а не отсутствие: (%q, %v)", v, err)
	}
	if v, err := Lookup(chkEnv, "MISSING", 42); v != 42 || err != nil {
		t.Fatalf("нет ключа — значение по умолчанию: (%v, %v)", v, err)
	}
}

func TestLookupDuration(t *testing.T) {
	v, err := Lookup(chkEnv, "TIMEOUT", 5*time.Second)
	if v != 90*time.Second || err != nil {
		t.Fatalf("TIMEOUT=1m30s → (%v, %v), ожидали 1m30s — Duration разбирают time.ParseDuration, а не как int64", v, err)
	}
}

func TestLookupTextUnmarshaler(t *testing.T) {
	if v, err := Lookup(chkEnv, "LEVEL", chkLevel(1)); v != 2 || err != nil {
		t.Fatalf("LEVEL = (%v, %v), ожидали 2 через UnmarshalText", v, err)
	}
	if v, err := Lookup(chkEnv, "ADDR", netip.Addr{}); v != netip.MustParseAddr("10.0.0.7") || err != nil {
		t.Fatalf("ADDR (netip.Addr) = (%v, %v)", v, err)
	}
	if v, err := Lookup(chkEnv, "START", time.Time{}); !v.Equal(time.Date(2024, 3, 1, 10, 0, 0, 0, time.UTC)) || err != nil {
		t.Fatalf("START (time.Time) = (%v, %v)", v, err)
	}
	if v, err := Lookup(map[string]string{"LEVEL": "trace"}, "LEVEL", chkLevel(1)); err == nil || v != 1 || !strings.HasPrefix(err.Error(), "LEVEL: ") {
		t.Fatalf("неизвестный уровень: (%v, %v) — ожидали значение по умолчанию и ошибку с именем ключа", v, err)
	}
}

func TestLookupErrors(t *testing.T) {
	v, err := Lookup(chkEnv, "BAD_INT", 80)
	var ne *strconv.NumError
	if v != 80 || !errors.As(err, &ne) || !strings.HasPrefix(err.Error(), "BAD_INT: ") {
		t.Fatalf("BAD_INT=80a → (%v, %v); ожидали 80 и ошибку \"BAD_INT: …\" с *strconv.NumError внутри", v, err)
	}
	if _, err := Lookup(chkEnv, "EMPTY", 1); err == nil {
		t.Fatalf("пустая строка для int должна быть ошибкой")
	}
	if _, err := Lookup(chkEnv, "YES", false); err == nil {
		t.Fatalf("yes — не bool для strconv.ParseBool")
	}
}

func TestLookupUnsupported(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("паника на неподдерживаемом типе: %v", r)
			}
		}()
		if v, err := Lookup(chkEnv, "PORT", []int{1}); !errors.Is(err, ErrUnsupported) || !reflect.DeepEqual(v, []int{1}) {
			t.Fatalf("[]int: (%v, %v), ожидали значение по умолчанию и ErrUnsupported", v, err)
		}
		if v, err := Lookup(chkEnv, "PORT", chkPort(80)); !errors.Is(err, ErrUnsupported) || v != 80 {
			t.Fatalf("type Port int: (%v, %v) — по контракту свой тип не поддерживается", v, err)
		}
		if v, err := Lookup(chkEnv, "MISSING", []int{1}); err != nil || len(v) != 1 {
			t.Fatalf("нет ключа — значение по умолчанию без ошибки для любого типа: (%v, %v)", v, err)
		}
	}()
}
