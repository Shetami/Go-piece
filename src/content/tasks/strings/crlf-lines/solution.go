package main

import (
	"bufio"
	"fmt"
	"strings"
)

func main() {
	// Файл пришёл из Windows: строки кончаются на \r\n.
	data := "admin\r\nguest\r\nroot\r\n"
	allowed := map[string]bool{"admin": true, "root": true}

	// bufio.ScanLines отрезает и \n, и \r перед ним.
	sc := bufio.NewScanner(strings.NewReader(data))
	for sc.Scan() {
		name := sc.Text()
		fmt.Printf("%q: %v\n", name, allowed[name])
	}
}
