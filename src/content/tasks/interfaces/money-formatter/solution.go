package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Money — сумма в минорных единицах (копейки, центы) и код валюты.
type Money struct {
	Minor int64
	Cur   string
}

// String: "12.50 RUB", "-0.05 USD", "0.00 EUR". Работает для любого int64.
func (m Money) String() string {
	return m.amount(false) + " " + m.Cur
}

func (m Money) amount(plus bool) string {
	neg := m.Minor < 0
	// Модуль в uint64: -MinInt64 в int64 не помещается.
	u := uint64(m.Minor)
	if neg {
		u = -u
	}
	s := strconv.FormatUint(u/100, 10) + "." + fmt.Sprintf("%02d", u%100)
	switch {
	case neg:
		return "-" + s
	case plus:
		return "+" + s
	}
	return s
}

// Format делает Money fmt.Formatter:
//   - %v и %s — как String; ширина работает (%12v — выравнивание вправо,
//     %-12v — влево);
//   - %+v — знак всегда: "+12.50 RUB", "+0.00 RUB", "-1.00 RUB";
//   - %d — только минорные единицы: "1250";
//   - прочие глаголы — "%!x(Money=12.50 RUB)" (x — сам глагол).
func (m Money) Format(f fmt.State, verb rune) {
	var s string
	switch verb {
	case 'v', 's':
		// Нельзя fmt.Fprintf(f, "%v", m): это снова вызовет Format — бесконечная рекурсия.
		s = m.amount(verb == 'v' && f.Flag('+')) + " " + m.Cur
	case 'd':
		s = strconv.FormatInt(m.Minor, 10)
	default:
		s = "%!" + string(verb) + "(Money=" + m.String() + ")"
	}
	if w, ok := f.Width(); ok && len(s) < w {
		pad := strings.Repeat(" ", w-len(s))
		if f.Flag('-') {
			s += pad
		} else {
			s = pad + s
		}
	}
	f.Write([]byte(s))
}

var ErrCurrencyMismatch = errors.New("currency mismatch")

// CurrencyError — попытка сложить разные валюты.
type CurrencyError struct{ Want, Got string }

func (e *CurrencyError) Error() string {
	return "currency mismatch: want " + e.Want + ", got " + e.Got
}

// Is делает errors.Is(err, ErrCurrencyMismatch) истинным для *CurrencyError.
func (e *CurrencyError) Is(target error) bool { return target == ErrCurrencyMismatch }

// Sum складывает суммы одной валюты. Пустой список — Money{}, nil.
// Разные валюты — *CurrencyError{Want: валюта первой, Got: первая чужая}.
func Sum(ms ...Money) (Money, error) {
	if len(ms) == 0 {
		return Money{}, nil
	}
	total := Money{Cur: ms[0].Cur}
	for _, m := range ms {
		if m.Cur != total.Cur {
			return Money{}, &CurrencyError{Want: total.Cur, Got: m.Cur}
		}
		total.Minor += m.Minor
	}
	return total, nil
}
