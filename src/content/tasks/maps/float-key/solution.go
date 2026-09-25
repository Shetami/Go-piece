package main

import (
	"fmt"
	"math"
)

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

	// Округляем до двух знаков: 0.1 + 0.2 в двоичной дроби — не ровно 0.3.
	gift, ok := bonus[math.Round(total*100)/100]
	fmt.Println(total, gift, ok)
}
