package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ErrInvalidDuration оборачивают все ошибки ParseDuration.
var ErrInvalidDuration = errors.New("некорректная длительность")

// units — от крупной единицы к мелкой. "ms" стоит раньше "m": при разборе
// единицы сначала пробуем более длинную.
var units = []struct {
	name string
	d    time.Duration
}{
	{"w", 7 * 24 * time.Hour},
	{"d", 24 * time.Hour},
	{"h", time.Hour},
	{"ms", time.Millisecond},
	{"m", time.Minute},
	{"s", time.Second},
}

// rank — место единицы в порядке «от крупной к мелкой».
var rank = map[string]int{"w": 0, "d": 1, "h": 2, "m": 3, "s": 4, "ms": 5}

// ParseDuration разбирает длительность вида "1w2d", "36h", "1d 12h30m",
// "-1h30m", "250ms".
//
// Правила:
//   - строка — одна или несколько компонент «целое число + единица»;
//     единицы: w (7 дней), d (24 часа), h, m, s, ms;
//   - компоненты идут строго от крупной единицы к мелкой, каждая не больше
//     одного раза ("1h2h" и "5m1h" — ошибки);
//   - между компонентами допустимы пробелы, пробелы по краям — тоже;
//     внутри компоненты (между числом и единицей) пробелов нет;
//   - в самом начале может стоять один '-', он относится ко всей длительности;
//   - "0" без единицы допустим и равен нулю, любое другое число без
//     единицы — ошибка; пустая строка — ошибка;
//   - если результат не помещается в time.Duration — ошибка.
//
// Все ошибки оборачивают ErrInvalidDuration и упоминают исходную строку.
func ParseDuration(s string) (time.Duration, error) {
	orig := s
	bad := func(why string) (time.Duration, error) {
		return 0, fmt.Errorf("%w %q: %s", ErrInvalidDuration, orig, why)
	}
	s = strings.TrimSpace(s)
	neg := false
	if rest, ok := strings.CutPrefix(s, "-"); ok {
		neg, s = true, rest
	}
	if s == "0" {
		return 0, nil
	}
	if s == "" {
		return bad("пустая строка")
	}
	var total time.Duration
	last := -1
	for s != "" {
		// Число.
		i := 0
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		if i == 0 {
			return bad("ожидали число")
		}
		n, err := strconv.ParseInt(s[:i], 10, 64)
		if err != nil {
			return bad("слишком большое число")
		}
		s = s[i:]
		// Единица: сначала длинные ("ms" раньше "m").
		var unit time.Duration
		var name string
		for _, u := range units {
			if strings.HasPrefix(s, u.name) {
				unit, name = u.d, u.name
				break
			}
		}
		if name == "" {
			return bad("нет единицы измерения")
		}
		if rank[name] <= last {
			return bad("единицы должны идти от крупной к мелкой без повторов")
		}
		last = rank[name]
		s = strings.TrimLeft(s[len(name):], " ")
		// Переполнение проверяем до умножения и до сложения.
		if n > (math.MaxInt64-int64(total))/int64(unit) {
			return bad("переполнение")
		}
		total += time.Duration(n) * unit
	}
	if neg {
		total = -total
	}
	return total, nil
}
