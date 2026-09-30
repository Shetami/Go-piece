package main

import "errors"

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
	// ваш код
	return ""
}

func (e *TransitionError) Is(target error) bool {
	// ваш код
	return false
}

// Fire применяет событие:
//   - ошибка On — вернуть её; состояние и история не меняются;
//   - On вернул nil — ErrNilState, ничего не меняется;
//   - новое состояние с тем же Name — переход «на месте»: хуков нет,
//     история не растёт;
//   - иначе: Exit у старого (если Exiter), смена o.State, запись в History,
//     Enter у нового (если Enterer) — именно в таком порядке.
func (o *Order) Fire(e Event) error {
	// ваш код
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
//
// Типы состояний объявите сами.

// NewOrder — заказ в состоянии new, History = ["new"].
func NewOrder(id string) *Order {
	// ваш код
	return &Order{ID: id}
}
