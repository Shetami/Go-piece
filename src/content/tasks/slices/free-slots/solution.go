package main

import (
	"cmp"
	"slices"
)

// Interval — полуоткрытый промежуток [Start, End) в минутах от полуночи.
type Interval struct{ Start, End int }

// FreeSlots возвращает свободные промежутки внутри рабочего дня day длиной
// не меньше minLen, по возрастанию. Пустых промежутков в ответе не бывает.
// busy — занятые интервалы в любом порядке: могут перекрываться, касаться,
// вкладываться друг в друга и выходить за границы дня. Интервал с
// Start >= End пустой и игнорируется. busy не меняется.
func FreeSlots(busy []Interval, day Interval, minLen int) []Interval {
	minLen = max(minLen, 1)
	b := slices.Clone(busy) // сортируем копию: вход чужой
	slices.SortFunc(b, func(x, y Interval) int { return cmp.Compare(x.Start, y.Start) })

	var out []Interval
	cursor := day.Start // всё левее cursor уже занято или разобрано
	for _, iv := range b {
		if iv.Start >= iv.End {
			continue
		}
		if iv.Start >= day.End {
			break
		}
		if iv.Start-cursor >= minLen {
			out = append(out, Interval{cursor, iv.Start})
		}
		// max, а не iv.End: вложенный интервал не должен отодвигать курсор назад.
		cursor = max(cursor, iv.End)
	}
	if day.End-cursor >= minLen {
		out = append(out, Interval{cursor, day.End})
	}
	return out
}
