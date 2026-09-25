package main

import "fmt"

// Бонус к заказу по сумме корзины в тысячах рублей.
var bonus = map[float64]string{
	0.3: "бесплатная доставка",
	0.5: "подарок",
}

func main() {
	cart := []float64{0.1, 0.2}

	var total float64
	for _, price := range cart {
		total += price
	}

	gift, ok := bonus[total]
	fmt.Println(total, gift, ok)
}
