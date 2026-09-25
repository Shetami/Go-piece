package main

import "fmt"

type Event struct {
	User   string
	Action string
}

// dedupKey — ключ, по которому одинаковые события считаются одним.
func dedupKey(e Event) any {
	return &e
}

func main() {
	events := []Event{
		{"аня", "вход"},
		{"боря", "вход"},
		{"аня", "вход"},
		{"аня", "вход"},
	}

	seen := map[any]bool{}
	for _, e := range events {
		seen[dedupKey(e)] = true
	}
	fmt.Println("уникальных событий:", len(seen))
}
