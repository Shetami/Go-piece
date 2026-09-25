package main

import (
	"fmt"
	"sync/atomic"
)

func try(name string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(name, "паника:", r)
		}
	}()
	f()
	fmt.Println(name, "ок")
}

func main() {
	var cfg atomic.Value
	fmt.Println(cfg.Load())

	try("строка", func() { cfg.Store("конфиг v1") })
	try("число", func() { cfg.Store(42) })
	try("nil", func() { cfg.Store(nil) })

	old := cfg.Swap("конфиг v2")
	fmt.Println(old, "→", cfg.Load())
}
