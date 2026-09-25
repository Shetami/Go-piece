package main

import "fmt"

type Config struct {
	Rate float64 // рублей за доллар
}

// pricer возвращает функцию пересчёта цены заказа в доллары.
// Курс фиксируется в момент оформления заказа.
func pricer(cfg *Config) func(rub float64) float64 {
	return func(rub float64) float64 {
		return rub / cfg.Rate
	}
}

func main() {
	cfg := &Config{Rate: 90}

	orderPrice := pricer(cfg) // заказ оформлен по курсу 90

	cfg.Rate = 100 // курс обновился
	fmt.Printf("%.2f\n", orderPrice(900))
}
