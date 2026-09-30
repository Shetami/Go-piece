package main

// chkID — Marshaler на значении.
type chkID int

func (id chkID) MarshalText() (string, error) {
	if id < 0 {
		return "", errors.New("negative id")
	}
	return fmt.Sprintf(`"id-%d"`, int(id)), nil
}

// chkLevel — и error, и Stringer: побеждает error.
type chkLevel int

func (l chkLevel) Error() string  { return "level error" }
func (l chkLevel) String() string { return "level string" }

type chkColor struct{ name string }

func (c chkColor) String() string { return c.name }

// chkBoth — и Marshaler, и error: побеждает Marshaler.
type chkBoth struct{}

func (chkBoth) MarshalText() (string, error) { return "true", nil }
func (chkBoth) Error() string                { return "both" }

func chkEnc(t *testing.T, v any, want string) {
	t.Helper()
	got, err := Encode(v)
	if err != nil || got != want {
		t.Fatalf("Encode(%#v) = %q, %v; ожидали %q", v, got, err, want)
	}
}

func TestEncodeBasic(t *testing.T) {
	chkEnc(t, nil, "null")
	chkEnc(t, true, "true")
	chkEnc(t, 42, "42")
	chkEnc(t, int64(-7), "-7")
	chkEnc(t, 0.5, "0.5")
	chkEnc(t, "a\"b", `"a\"b"`)
	chkEnc(t, []any{1, "x", nil, []any{}}, `[1,"x",null,[]]`)
	chkEnc(t, map[string]any{"b": 2, "a": []any{true}, "c": map[string]any{}}, `{"a":[true],"b":2,"c":{}}`)
}

func TestEncodeInterfaces(t *testing.T) {
	chkEnc(t, chkID(7), `"id-7"`)
	chkEnc(t, errors.New("boom"), `"boom"`)
	chkEnc(t, fmt.Errorf("wrap: %w", io.EOF), `"wrap: EOF"`)
	chkEnc(t, chkColor{"red"}, `"red"`)
	chkEnc(t, chkLevel(1), `"level error"`)
	chkEnc(t, chkBoth{}, "true")
	chkEnc(t, map[string]any{"id": chkID(1), "c": &chkColor{"blue"}}, `{"c":"blue","id":"id-1"}`)
}

func TestEncodeNilPointers(t *testing.T) {
	var id *chkID
	var c *chkColor
	var n *int
	chkEnc(t, id, "null")
	chkEnc(t, c, "null")
	chkEnc(t, []any{n, id}, "[null,null]")
}

func TestEncodeErrors(t *testing.T) {
	_, err := Encode([]any{1, map[string]any{"x": chkID(-1)}})
	if err == nil || !strings.Contains(err.Error(), "negative id") {
		t.Fatalf("ошибка MarshalText во вложенном элементе: %v", err)
	}
	type point struct{ X int }
	out, err := Encode(map[string]any{"p": point{1}})
	if out != "" || !errors.Is(err, ErrUnsupported) || !strings.Contains(err.Error(), "point") {
		t.Fatalf("неподдерживаемый тип: %q, %v; ожидали пустую строку и ErrUnsupported с именем типа", out, err)
	}
	if _, err := Encode([]int{1}); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("[]int — не []any: %v, ожидали ErrUnsupported", err)
	}
}
