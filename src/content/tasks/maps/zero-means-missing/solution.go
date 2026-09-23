package main

import "fmt"

var prices = map[string]int{"кофе": 200, "вода": 0, "чай": 150}

func describe(item string) string {
	// ok отличает «в меню, но бесплатно» от «нет в меню».
	price, ok := prices[item]
	if !ok {
		return item + ": нет в меню"
	}
	return fmt.Sprintf("%s: %d ₽", item, price)
}

func main() {
	for _, item := range []string{"кофе", "вода", "какао"} {
		fmt.Println(describe(item))
	}
}
