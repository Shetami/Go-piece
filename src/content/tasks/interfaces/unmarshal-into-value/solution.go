package main

import (
	"encoding/json"
	"fmt"
)

type Config struct {
	Port  int  `json:"port"`
	Debug bool `json:"debug"`
}

func main() {
	var cfg Config
	// Unmarshal должен писать в нашу переменную — передаём её адрес.
	err := json.Unmarshal([]byte(`{"port": 8080, "debug": true}`), &cfg)
	fmt.Println(err)
	fmt.Printf("%+v\n", cfg)
}
