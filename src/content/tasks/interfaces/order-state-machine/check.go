package main

// Тестовые состояния для проверки самого Fire.
type chkA struct{ log *[]string }
type chkB struct{ log *[]string }
type chkNil struct{}

func (s chkA) Name() string { return "A" }
func (s chkA) On(e Event, o *Order) (State, error) {
	switch e {
	case "go":
		return chkB{s.log}, nil
	case "stay":
		return chkA{s.log}, nil
	}
	return nil, errors.New("nope")
}
func (s chkA) Exit(o *Order)  { *s.log = append(*s.log, "exit A, state="+o.State.Name()) }
func (s chkA) Enter(o *Order) { *s.log = append(*s.log, "enter A") }

func (s chkB) Name() string                        { return "B" }
func (s chkB) On(e Event, o *Order) (State, error) { return chkNil{}, nil }
func (s chkB) Enter(o *Order)                      { *s.log = append(*s.log, "enter B, state="+o.State.Name()) }

func (chkNil) Name() string                    { return "N" }
func (chkNil) On(Event, *Order) (State, error) { return nil, nil }

func TestFireHooks(t *testing.T) {
	var log []string
	o := &Order{State: chkA{&log}, History: []string{"A"}}
	if err := o.Fire("stay"); err != nil || len(log) != 0 || len(o.History) != 1 {
		t.Fatalf("переход в то же состояние: err=%v, хуки %v, история %v; ожидали без хуков и записи", err, log, o.History)
	}
	if err := o.Fire("go"); err != nil {
		t.Fatal(err)
	}
	want := "exit A, state=A; enter B, state=B"
	if got := strings.Join(log, "; "); got != want {
		t.Fatalf("хуки: %q, ожидали %q", got, want)
	}
	if strings.Join(o.History, ",") != "A,B" {
		t.Fatalf("история %v, ожидали [A B]", o.History)
	}
	o = &Order{State: chkNil{}, History: []string{"N"}}
	if err := o.Fire("x"); !errors.Is(err, ErrNilState) || o.State == nil {
		t.Fatalf("On вернул (nil, nil): %v, состояние %v; ожидали ErrNilState и прежнее состояние", err, o.State)
	}
	o = &Order{State: chkA{&log}, History: []string{"A"}}
	if err := o.Fire("bad"); err == nil || o.State.Name() != "A" || len(o.History) != 1 {
		t.Fatalf("ошибка On: %v, состояние %s, история %v", err, o.State.Name(), o.History)
	}
}

func chkFire(t *testing.T, o *Order, events ...Event) {
	t.Helper()
	for _, e := range events {
		if err := o.Fire(e); err != nil {
			t.Fatalf("Fire(%s) = %v", e, err)
		}
	}
}

func TestOrderHappyPath(t *testing.T) {
	o := NewOrder("42")
	if o.State == nil || o.State.Name() != "new" {
		t.Fatal("NewOrder должен создавать заказ в состоянии new")
	}
	chkFire(t, o, Pay, Ship, Deliver)
	if got := strings.Join(o.History, ","); got != "new,paid,shipped,delivered" || o.Refunded {
		t.Fatalf("история %s, возврат %v", got, o.Refunded)
	}
}

func TestOrderInvalid(t *testing.T) {
	o := NewOrder("1")
	chkFire(t, o, Pay, Ship)
	err := o.Fire(Cancel)
	var te *TransitionError
	if !errors.Is(err, ErrInvalidTransition) || !errors.As(err, &te) || te.State != "shipped" || te.Event != Cancel {
		t.Fatalf("отмена отправленного: %v; ожидали *TransitionError{shipped, cancel}", err)
	}
	if err.Error() != "cannot cancel in shipped" {
		t.Fatalf("текст ошибки %q", err)
	}
	if o.State.Name() != "shipped" || len(o.History) != 3 {
		t.Fatalf("после ошибки: состояние %s, история %v", o.State.Name(), o.History)
	}
	if err := NewOrder("2").Fire(Deliver); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("deliver в new: %v", err)
	}
}

func TestOrderCancel(t *testing.T) {
	o := NewOrder("1")
	chkFire(t, o, Cancel)
	if o.State.Name() != "cancelled" || o.Refunded {
		t.Fatalf("отмена неоплаченного: %s, возврат %v; ожидали cancelled без возврата", o.State.Name(), o.Refunded)
	}
	if err := o.Fire(Pay); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("оплата отменённого: %v", err)
	}
	o = NewOrder("2")
	chkFire(t, o, Pay, Cancel)
	if o.State.Name() != "refunded" || !o.Refunded || strings.Join(o.History, ",") != "new,paid,refunded" {
		t.Fatalf("отмена оплаченного: %s, возврат %v, история %v", o.State.Name(), o.Refunded, o.History)
	}
}
