package main

import (
	"errors"
	"fmt"
)

type Config struct {
	Raw string
	Env string
}

func readFile(name string) (string, error) {
	return "", errors.New("open " + name + ": no such file")
}

func lookupEnv(key string) (string, error) {
	return "prod", nil
}

func loadConfig() (Config, error) {
	var cfg Config
	data, err := readFile("app.yaml")
	if err != nil {
		// Проверяем сразу: следующий вызов перезапишет err.
		return Config{}, fmt.Errorf("конфиг: %w", err)
	}
	cfg.Raw = data
	cfg.Env, err = lookupEnv("APP_ENV")
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func main() {
	cfg, err := loadConfig()
	fmt.Printf("%+v %v\n", cfg, err)
}
