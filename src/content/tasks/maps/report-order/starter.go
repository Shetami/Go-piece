package main

import (
	"fmt"
	"maps"
	"slices"
)

func main() {
	sales := map[string]int{"март": 120, "январь": 100, "февраль": 90, "апрель": 150}

	// Отчёт: месяцы по убыванию продаж.
	for _, month := range slices.Sorted(maps.Keys(sales)) {
		fmt.Printf("%s: %d\n", month, sales[month])
	}
}
