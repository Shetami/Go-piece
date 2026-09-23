package main

import (
	"fmt"
	"strconv"
	"strings"
)

func main() {
	input := "10, 20 ,30"
	sum := 0
	for _, part := range strings.Split(input, ",") {
		// Atoi не прощает пробелов — срезаем их по краям.
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			fmt.Println("пропускаю:", err)
			continue
		}
		sum += n
	}
	fmt.Println("сумма:", sum)
}
