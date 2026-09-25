package main

import (
	"fmt"
	"slices"
)

// newBlocklist готовит список для быстрого поиска:
// сортирует его, чтобы искать двоичным поиском.
func newBlocklist(hosts ...string) []string {
	list := slices.Clone(hosts)
	// Порядок сортировки обязан совпадать с тем, которым ищет BinarySearch.
	slices.Sort(list)
	return list
}

func blocked(list []string, host string) bool {
	_, ok := slices.BinarySearch(list, host)
	return ok
}

func main() {
	list := newBlocklist("spam.ru", "ads.example.com", "track.io", "x.co", "bad.net")
	for _, h := range []string{"ads.example.com", "bad.net", "x.co", "track.io", "go.dev"} {
		fmt.Println(h, blocked(list, h))
	}
}
