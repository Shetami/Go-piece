package main

import (
	"fmt"
	"sync"
)

var (
	once sync.Once
	cfg  map[string]string
)

// Config лениво собирает конфиг при первом обращении.
func Config() map[string]string {
	once.Do(func() {
		c := map[string]string{"host": "localhost"}
		// Внутри инициализации читаем то, что собираем сейчас, а не идём
		// снова через Config: повторный once.Do ждал бы сам себя.
		c["addr"] = c["host"] + ":8080"
		cfg = c
	})
	return cfg
}

func Get(key string) string {
	return Config()[key]
}

func main() {
	fmt.Println(Config()["addr"])
}
