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
		cfg = map[string]string{"host": "localhost"}
		cfg["addr"] = Get("host") + ":8080"
	})
	return cfg
}

func Get(key string) string {
	return Config()[key]
}

func main() {
	fmt.Println(Config()["addr"])
}
