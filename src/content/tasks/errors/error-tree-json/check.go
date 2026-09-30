package main

type chkCoded struct{ code, msg string }

func (e *chkCoded) Error() string { return e.msg }
func (e *chkCoded) Code() string  { return e.code }

type chkLoop struct{}

func (e *chkLoop) Error() string { return "петля" }
func (e *chkLoop) Unwrap() error { return e }

type chkMulti []error

func (m chkMulti) Error() string   { return "мульти" }
func (m chkMulti) Unwrap() []error { return m }

func chkJSON(t *testing.T, err error) string {
	t.Helper()
	b, e := json.Marshal(Encode(err))
	if e != nil {
		t.Fatalf("json.Marshal: %v", e)
	}
	return string(b)
}

func TestEncodeNil(t *testing.T) {
	if Encode(nil) != nil {
		t.Fatal("Encode(nil) должен вернуть nil")
	}
}

func TestEncodeChain(t *testing.T) {
	err := fmt.Errorf("заказ 7: %w", &chkCoded{"db.timeout", "база не ответила"})
	want := `{"msg":"заказ 7: база не ответила","causes":[{"msg":"база не ответила","code":"db.timeout"}]}`
	if got := chkJSON(t, err); got != want {
		t.Fatalf("получили\n%s\nожидали\n%s\n(код — только у звена, которое само его реализует)", got, want)
	}
}

func TestEncodeJoin(t *testing.T) {
	err := fmt.Errorf("сохранить: %w", errors.Join(&chkCoded{"db", "нет соединения"}, fmt.Errorf("кэш: %w", io.EOF)))
	want := `{"msg":"сохранить: нет соединения\nкэш: EOF","causes":[{"msg":"нет соединения\nкэш: EOF","causes":[{"msg":"нет соединения","code":"db"},{"msg":"кэш: EOF","causes":[{"msg":"EOF"}]}]}]}`
	if got := chkJSON(t, err); got != want {
		t.Fatalf("получили\n%s\nожидали\n%s", got, want)
	}
	two := fmt.Errorf("a: %w, b: %w", io.EOF, io.ErrClosedPipe)
	if n := Encode(two); n == nil || len(n.Causes) != 2 {
		t.Fatalf("у fmt.Errorf с двумя %%w два ребёнка, а получили %+v", Encode(two))
	}
}

func TestEncodeSkipsNil(t *testing.T) {
	n := Encode(chkMulti{io.EOF, nil, io.ErrUnexpectedEOF})
	if n == nil || len(n.Causes) != 2 || n.Causes[1].Msg != io.ErrUnexpectedEOF.Error() {
		t.Fatalf("nil-элементы Unwrap() []error надо пропускать, а получили %+v", n)
	}
}

func TestEncodeDepth(t *testing.T) {
	done := make(chan *Node, 1)
	go func() { done <- Encode(&chkLoop{}) }()
	var n *Node
	select {
	case n = <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Encode зациклился на ошибке, которая разворачивается сама в себя")
	}
	depth := 0
	for n != nil {
		depth++
		if len(n.Causes) == 0 {
			break
		}
		n = &n.Causes[0]
	}
	if depth != MaxDepth {
		t.Fatalf("глубина дерева %d, ожидали MaxDepth = %d", depth, MaxDepth)
	}
}
