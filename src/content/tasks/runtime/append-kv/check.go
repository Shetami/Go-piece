package main

func TestAppendKVFormat(t *testing.T) {
	cases := []struct {
		key  string
		val  any
		want string
	}{
		{"port", 8080, `port=8080`},
		{"delta", -3, `delta=-3`},
		{"user", "аня", `user="аня"`},
		{"msg", "a \"b\"\n", `msg="a \"b\"\n"`},
		{"debug", true, `debug=true`},
		{"ratio", 0.5, `ratio=?`},
	}
	for _, c := range cases {
		if got := string(AppendKV(nil, c.key, c.val)); got != c.want {
			t.Fatalf("AppendKV(%q, %#v) = %s, ожидали %s", c.key, c.val, got, c.want)
		}
	}
}

func TestAppendKVAppends(t *testing.T) {
	b := []byte("lvl=info ")
	b = AppendKV(b, "code", 200)
	if string(b) != "lvl=info code=200" {
		t.Fatalf("получили %q — AppendKV должна дописывать, а не перезаписывать", b)
	}
}

var sinkBuf []byte

func TestAppendKVNoAllocs(t *testing.T) {
	buf := make([]byte, 0, 256)
	for _, val := range []any{123456, "значение с пробелом", true} {
		allocs := testing.AllocsPerRun(100, func() {
			sinkBuf = AppendKV(buf[:0], "key", val)
		})
		if allocs != 0 {
			t.Fatalf("для %T: %.0f выделений на вызов при достаточном буфере, ожидали 0", val, allocs)
		}
	}
}
