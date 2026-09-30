package main

import (
	"errors"
	"fmt"
)

type Event string

const (
	Pay     Event = "pay"
	Ship    Event = "ship"
	Deliver Event = "deliver"
	Cancel  Event = "cancel"
)

// State — состояние заказа. On решает, куда перейти по событию.
type State interface {
	Name() string
	On(e Event, o *Order) (State, error)
}

// Опциональные хуки состояния.
type Enterer interface{ Enter(o *Order) }
type Exiter interface{ Exit(o *Order) }

type Order struct {
	ID       string
	State    State
	History  []string // имена состояний, начиная с начального
	Refunded bool
}

var (
	ErrInvalidTransition = errors.New("invalid transition")
	ErrNilState          = errors.New("state returned nil")
)

// TransitionError — событие недопустимо в текущем состоянии.
// Текст: "cannot <event> in <state>"; errors.Is(err, ErrInvalidTransition) — true.
type TransitionError struct {
	State string
	Event Event
}

func (e *TransitionError) Error() string {
	return fmt.Sprintf("cannot %s in %s", e.Event, e.State)
}

func (e *TransitionError) Is(target error) bool { return target == ErrInvalidTransition }

// Fire применяет событие:
//   - ошибка On — вернуть её; состояние и история не меняются;
//   - On вернул nil — ErrNilState, ничего не меняется;
//   - новое состояние с тем же Name — переход «на месте»: хуков нет,
//     история не растёт;
//   - иначе: Exit у старого (если Exiter), смена o.State, запись в History,
//     Enter у нового (если Enterer) — именно в таком порядке.
func (o *Order) Fire(e Event) error {
	next, err := o.State.On(e, o)
	if err != nil {
		return err
	}
	if next == nil {
		return ErrNilState
	}
	if next.Name() == o.State.Name() {
		o.State = next
		return nil
	}
	if x, ok := o.State.(Exiter); ok {
		x.Exit(o)
	}
	o.State = next
	o.History = append(o.History, next.Name())
	if en, ok := next.(Enterer); ok {
		en.Enter(o) // хук видит уже новое состояние
	}
	return nil
}

// Состояния заказа:
//
//	new     --pay-->     paid      --ship-->   shipped --deliver--> delivered
//	new     --cancel-->  cancelled
//	paid    --cancel-->  refunded  (Enter ставит o.Refunded = true)
//
// delivered, cancelled, refunded — конечные. Любое недопустимое событие —
// *TransitionError.
type (
	stNew       struct{}
	stPaid      struct{}
	stShipped   struct{}
	stDelivered struct{}
	stCancelled struct{}
	stRefunded  struct{}
)

func invalid(s State, e Event) error { return &TransitionError{State: s.Name(), Event: e} }

func (s stNew) Name() string { return "new" }
func (s stNew) On(e Event, _ *Order) (State, error) {
	switch e {
	case Pay:
		return stPaid{}, nil
	case Cancel:
		return stCancelled{}, nil
	}
	return nil, invalid(s, e)
}

func (s stPaid) Name() string { return "paid" }
func (s stPaid) On(e Event, _ *Order) (State, error) {
	switch e {
	case Ship:
		return stShipped{}, nil
	case Cancel:
		return stRefunded{}, nil
	}
	return nil, invalid(s, e)
}

func (s stShipped) Name() string { return "shipped" }
func (s stShipped) On(e Event, _ *Order) (State, error) {
	if e == Deliver {
		return stDelivered{}, nil
	}
	return nil, invalid(s, e)
}

func (s stDelivered) Name() string                        { return "delivered" }
func (s stDelivered) On(e Event, _ *Order) (State, error) { return nil, invalid(s, e) }
func (s stCancelled) Name() string                        { return "cancelled" }
func (s stCancelled) On(e Event, _ *Order) (State, error) { return nil, invalid(s, e) }
func (s stRefunded) Name() string                         { return "refunded" }
func (s stRefunded) On(e Event, _ *Order) (State, error)  { return nil, invalid(s, e) }
func (s stRefunded) Enter(o *Order)                       { o.Refunded = true }

// NewOrder — заказ в состоянии new, History = ["new"].
func NewOrder(id string) *Order {
	return &Order{ID: id, State: stNew{}, History: []string{"new"}}
}
