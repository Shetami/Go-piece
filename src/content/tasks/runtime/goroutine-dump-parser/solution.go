package main

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Goroutine — одна горутина из дампа runtime.Stack(buf, true) или из
// вывода паники/SIGQUIT.
type Goroutine struct {
	ID        int
	State     string        // "chan receive", "select", "running", "chan receive (nil chan)"…
	Wait      time.Duration // "…, 5 minutes]" → 5*time.Minute; нет — 0
	Top       string        // функция верхнего кадра без аргументов: "main.(*Server).handle"
	CreatedBy string        // "main.main" из "created by main.main in goroutine 1"; нет — ""
}

// ParseStacks разбирает дамп. Блоки горутин разделены пустой строкой;
// заголовок блока: "goroutine <id> [<состояние>[, <N> minutes][, locked to thread]]:".
// Некорректный заголовок — ошибка.
func ParseStacks(dump string) ([]Goroutine, error) {
	var out []Goroutine
	for _, block := range strings.Split(strings.TrimSpace(dump), "\n\n") {
		lines := strings.Split(strings.TrimSpace(block), "\n")
		if len(lines) == 0 || lines[0] == "" {
			continue
		}
		g, err := parseHeader(lines[0])
		if err != nil {
			return nil, err
		}
		for _, l := range lines[1:] {
			if strings.HasPrefix(l, "\t") {
				continue // строка "файл:строка +0x…"
			}
			if rest, ok := strings.CutPrefix(l, "created by "); ok {
				// Новый формат: "created by f in goroutine N", старый — без хвоста.
				g.CreatedBy, _, _ = strings.Cut(rest, " in goroutine ")
				continue
			}
			if g.Top == "" && !strings.HasPrefix(l, "...") {
				g.Top = funcName(l)
			}
		}
		out = append(out, g)
	}
	return out, nil
}

func parseHeader(h string) (Goroutine, error) {
	var g Goroutine
	rest, ok := strings.CutPrefix(h, "goroutine ")
	open := strings.IndexByte(rest, '[')
	if !ok || open < 0 || !strings.HasSuffix(rest, "]:") {
		return g, fmt.Errorf("некорректный заголовок %q", h)
	}
	id, err := strconv.Atoi(strings.TrimSpace(rest[:open]))
	if err != nil {
		return g, fmt.Errorf("некорректный id в %q", h)
	}
	g.ID = id
	// Внутри скобок: состояние, затем необязательные части через ", ".
	parts := strings.Split(rest[open+1:len(rest)-2], ", ")
	g.State = parts[0]
	for _, p := range parts[1:] {
		if n, ok := strings.CutSuffix(p, " minutes"); ok {
			if m, err := strconv.Atoi(n); err == nil {
				g.Wait = time.Duration(m) * time.Minute
			}
		}
	}
	return g, nil
}

// funcName отрезает аргументы: "main.(*S).f(0x1, ...)" → "main.(*S).f".
// Скобки есть и в имени метода, поэтому ищем ПОСЛЕДНЮЮ открывающую.
func funcName(line string) string {
	if strings.HasSuffix(line, ")") {
		if i := strings.LastIndexByte(line, '('); i > 0 {
			return line[:i]
		}
	}
	return line
}

// Group — горутины с одинаковыми State и Top.
type Group struct {
	State, Top string
	IDs        []int // по возрастанию
}

// GroupStacks группирует горутины. Порядок групп: по убыванию числа
// горутин, затем по State, затем по Top.
func GroupStacks(gs []Goroutine) []Group {
	type key struct{ state, top string }
	idx := map[key]int{}
	var out []Group
	for _, g := range gs {
		k := key{g.State, g.Top}
		i, ok := idx[k]
		if !ok {
			i = len(out)
			idx[k] = i
			out = append(out, Group{State: g.State, Top: g.Top})
		}
		out[i].IDs = append(out[i].IDs, g.ID)
	}
	for i := range out {
		slices.Sort(out[i].IDs)
	}
	slices.SortFunc(out, func(a, b Group) int {
		return cmp.Or(
			cmp.Compare(len(b.IDs), len(a.IDs)),
			cmp.Compare(a.State, b.State),
			cmp.Compare(a.Top, b.Top),
		)
	})
	return out
}
