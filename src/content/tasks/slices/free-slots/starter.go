package main

// Interval — полуоткрытый промежуток [Start, End) в минутах от полуночи.
type Interval struct{ Start, End int }

// FreeSlots возвращает свободные промежутки внутри рабочего дня day длиной
// не меньше minLen, по возрастанию. Пустых промежутков в ответе не бывает.
// busy — занятые интервалы в любом порядке: могут перекрываться, касаться,
// вкладываться друг в друга и выходить за границы дня. Интервал с
// Start >= End пустой и игнорируется. busy не меняется.
func FreeSlots(busy []Interval, day Interval, minLen int) []Interval {
	// ваш код
	return nil
}
