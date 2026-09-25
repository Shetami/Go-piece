package main

import "fmt"

type Product struct {
	Name  string
	Price int
}

// cheap возвращает указатели на товары дешевле limit,
// чтобы вызывающий мог поменять им цену прямо в каталоге.
func cheap(catalog []Product, limit int) []*Product {
	var out []*Product
	for _, p := range catalog {
		if p.Price < limit {
			out = append(out, &p)
		}
	}
	return out
}

func main() {
	catalog := []Product{{"чай", 90}, {"кофе", 350}, {"сахар", 60}}

	for _, p := range cheap(catalog, 100) {
		p.Price += 10
	}
	fmt.Println(catalog)
}
