package main

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Version — разобранная семантическая версия. Build в сравнении не участвует.
type Version struct {
	Major, Minor, Patch int
	Pre                 []string // идентификаторы pre-release: "rc.1" → ["rc", "1"]
	Build               string
}

// ParseVersion разбирает версию по SemVer 2.0.0: MAJOR.MINOR.PATCH,
// за ним необязательный pre-release после '-' и build после '+'.
// Допускается ведущая 'v' ("v1.2.3").
//
//   - MAJOR, MINOR, PATCH — неотрицательные целые без ведущих нулей ("01" — ошибка);
//   - pre-release — идентификаторы через точку, каждый непустой, из
//     [0-9A-Za-z-]; числовой идентификатор без ведущих нулей;
//   - build — идентификаторы через точку, непустые, из [0-9A-Za-z-].
//
// Текст ошибки содержит исходную строку в кавычках, как её печатает %q.
func ParseVersion(s string) (Version, error) {
	var v Version
	bad := func(why string) (Version, error) {
		return Version{}, fmt.Errorf("версия %q: %s", s, why)
	}
	rest := strings.TrimPrefix(s, "v")
	rest, build, hasBuild := strings.Cut(rest, "+")
	if hasBuild {
		if !validIdents(build, false) {
			return bad("некорректный build")
		}
		v.Build = build
	}
	core, pre, hasPre := strings.Cut(rest, "-") // первый '-': дальше '-' разрешён в идентификаторах
	if hasPre {
		if !validIdents(pre, true) {
			return bad("некорректный pre-release")
		}
		v.Pre = strings.Split(pre, ".")
	}
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return bad("ожидали MAJOR.MINOR.PATCH")
	}
	nums := [3]*int{&v.Major, &v.Minor, &v.Patch}
	for i, p := range parts {
		if !isNumeric(p) || (len(p) > 1 && p[0] == '0') {
			return bad("некорректное число " + strconv.Quote(p))
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return bad("слишком большое число")
		}
		*nums[i] = n
	}
	return v, nil
}

func validIdents(s string, pre bool) bool {
	for _, id := range strings.Split(s, ".") {
		if id == "" {
			return false
		}
		for _, c := range id {
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '-') {
				return false
			}
		}
		if pre && isNumeric(id) && len(id) > 1 && id[0] == '0' {
			return false
		}
	}
	return true
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// Compare возвращает -1, 0 или 1 по правилам приоритета SemVer.
func (v Version) Compare(w Version) int {
	if c := cmp.Compare(v.Major, w.Major); c != 0 {
		return c
	}
	if c := cmp.Compare(v.Minor, w.Minor); c != 0 {
		return c
	}
	if c := cmp.Compare(v.Patch, w.Patch); c != 0 {
		return c
	}
	// Релиз старше любого своего pre-release.
	switch {
	case len(v.Pre) == 0 && len(w.Pre) == 0:
		return 0
	case len(v.Pre) == 0:
		return 1
	case len(w.Pre) == 0:
		return -1
	}
	for i := 0; i < len(v.Pre) && i < len(w.Pre); i++ {
		if c := comparePreIdent(v.Pre[i], w.Pre[i]); c != 0 {
			return c
		}
	}
	// Все общие идентификаторы равны — меньше та, где их меньше.
	return cmp.Compare(len(v.Pre), len(w.Pre))
}

func comparePreIdent(a, b string) int {
	an, bn := isNumeric(a), isNumeric(b)
	switch {
	case an && bn:
		// Числа сравниваем как числа: сначала по длине (ведущих нулей нет),
		// потом лексикографически — без риска переполнения.
		if c := cmp.Compare(len(a), len(b)); c != 0 {
			return c
		}
		return strings.Compare(a, b)
	case an:
		return -1 // числовой идентификатор младше буквенного
	case bn:
		return 1
	}
	return strings.Compare(a, b) // ASCII-порядок: "RC" < "alpha"
}

// SortVersions возвращает новый слайс версий, отсортированный по
// возрастанию приоритета SemVer. Исходный слайс не меняется. Версии с
// равным приоритетом (отличаются только build) сохраняют исходный порядок.
// Если хоть одна строка некорректна, возвращает nil и ошибку, собранную
// через errors.Join из ошибок всех некорректных строк (в исходном порядке).
func SortVersions(vs []string) ([]string, error) {
	type item struct {
		s string
		v Version
	}
	items := make([]item, 0, len(vs))
	var errs []error
	for _, s := range vs {
		v, err := ParseVersion(s)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		items = append(items, item{s, v})
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	slices.SortStableFunc(items, func(a, b item) int { return a.v.Compare(b.v) })
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.s
	}
	return out, nil
}
