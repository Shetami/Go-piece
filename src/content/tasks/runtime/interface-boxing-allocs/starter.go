package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

var sink any

func main() {
	// len(os.Args) == 1: значения те же, но компилятор не знает их заранее
	// и не может подставить готовые константы.
	n := len(os.Args)
	small := 6 + n
	big := 999 + n
	s := strings.Repeat("го", n)
	p := &big

	fmt.Println("int 7:", testing.AllocsPerRun(100, func() { sink = small }))
	fmt.Println("int 1000:", testing.AllocsPerRun(100, func() { sink = big }))
	fmt.Println("string:", testing.AllocsPerRun(100, func() { sink = s }))
	fmt.Println("указатель:", testing.AllocsPerRun(100, func() { sink = p }))
}
