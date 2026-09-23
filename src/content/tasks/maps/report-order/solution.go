package main

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
)

func main() {
	sales := map[string]int{"март": 120, "январь": 100, "февраль": 90, "апрель": 150}

	// Отчёт: месяцы по убыванию продаж.
	// У мапы порядка нет — достаём ключи и сортируем их сами.
	months := slices.Collect(maps.Keys(sales))
	slices.SortFunc(months, func(a, b string) int {
		return cmp.Compare(sales[b], sales[a])
	})
	for _, month := range months {
		fmt.Printf("%s: %d\n", month, sales[month])
	}
}
