package main

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
)

var (
	// ErrSize — строка не похожа на размер.
	ErrSize = errors.New("некорректный размер")
	// ErrOverflow — размер не помещается в int64.
	ErrOverflow = errors.New("размер слишком большой")
)

var sizeUnits = map[string]int64{
	"": 1, "b": 1,
	"k": 1e3, "kb": 1e3, "m": 1e6, "mb": 1e6, "g": 1e9, "gb": 1e9, "t": 1e12, "tb": 1e12,
	"ki": 1 << 10, "kib": 1 << 10, "mi": 1 << 20, "mib": 1 << 20,
	"gi": 1 << 30, "gib": 1 << 30, "ti": 1 << 40, "tib": 1 << 40,
}

// ParseSize разбирает размер из конфига или флага: "512", "10MB",
// "1.5 GiB", "2k", "  3 KiB ".
//
//   - число: цифры, необязательно '.' и ещё цифры ("1.", ".5", "1,5",
//     "-1", "+1" — ошибки);
//   - единица без учёта регистра: B; K/KB, M/MB, G/GB, T/TB — степени
//     1000; Ki/KiB, Mi/MiB, Gi/GiB, Ti/TiB — степени 1024; без единицы — байты;
//   - между числом и единицей и по краям — сколько угодно пробелов;
//   - результат считается точно, без float: "1.1GB" — ровно 1100000000;
//     если байтов получается нецелое число ("0.5B", "1.0001KiB") — ErrSize;
//   - больше math.MaxInt64 — ErrOverflow; остальные ошибки — ErrSize.
//
// Ошибки оборачивают ErrSize или ErrOverflow и содержат исходную строку.
func ParseSize(s string) (int64, error) {
	bad := func(e error) (int64, error) { return 0, fmt.Errorf("%w: %q", e, s) }
	t := strings.TrimSpace(s)
	i := 0
	for i < len(t) && (t[i] >= '0' && t[i] <= '9' || t[i] == '.') {
		i++
	}
	num, unitStr := t[:i], strings.ToLower(strings.TrimSpace(t[i:]))
	intPart, frac, hasDot := strings.Cut(num, ".")
	if intPart == "" || (hasDot && (frac == "" || strings.Contains(frac, "."))) {
		return bad(ErrSize)
	}
	unit, ok := sizeUnits[unitStr]
	if !ok {
		return bad(ErrSize)
	}
	// Точная арифметика: 1.1 во float64 — это 1.100000000000000088…,
	// и 1.1GB превратилось бы в 1100000000.0000002.
	mant, _ := new(big.Int).SetString(intPart+frac, 10)
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(len(frac))), nil)
	total := new(big.Int).Mul(mant, big.NewInt(unit))
	q, r := new(big.Int).QuoRem(total, scale, new(big.Int))
	if r.Sign() != 0 {
		return bad(ErrSize) // нецелое число байт
	}
	if q.Cmp(big.NewInt(math.MaxInt64)) > 0 {
		return bad(ErrOverflow)
	}
	return q.Int64(), nil
}
