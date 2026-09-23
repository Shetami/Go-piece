package main

import (
	"fmt"
	"strings"
)

// baseName отрезает расширение .json.
func baseName(file string) string {
	return strings.TrimRight(file, ".json")
}

func main() {
	for _, f := range []string{"user.json", "config.json", "lesson.json"} {
		fmt.Println(baseName(f))
	}
}
