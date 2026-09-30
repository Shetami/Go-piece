package main

import (
	"errors"
	"fmt"
)

// Money — сумма в минорных единицах (копейки, центы) и код валюты.
type Money struct {
	Minor int64
	Cur   string
}

// String: "12.50 RUB", "-0.05 USD", "0.00 EUR". Работает для любого int64.
func (m Money) String() string {
	// ваш код
	return ""
}

// Format делает Money fmt.Formatter:
//   - %v и %s — как String; ширина работает (%12v — выравнивание вправо,
//     %-12v — влево);
//   - %+v — знак всегда: "+12.50 RUB", "+0.00 RUB", "-1.00 RUB";
//   - %d — только минорные единицы: "1250";
//   - прочие глаголы — "%!x(Money=12.50 RUB)" (x — сам глагол).
func (m Money) Format(f fmt.State, verb rune) {
	// ваш код
}

var ErrCurrencyMismatch = errors.New("currency mismatch")

// CurrencyError — попытка сложить разные валюты.
type CurrencyError struct{ Want, Got string }

func (e *CurrencyError) Error() string {
	// ваш код
	return ""
}

// Is делает errors.Is(err, ErrCurrencyMismatch) истинным для *CurrencyError.
func (e *CurrencyError) Is(target error) bool {
	// ваш код
	return false
}

// Sum складывает суммы одной валюты. Пустой список — Money{}, nil.
// Разные валюты — *CurrencyError{Want: валюта первой, Got: первая чужая}.
func Sum(ms ...Money) (Money, error) {
	// ваш код
	return Money{}, nil
}
