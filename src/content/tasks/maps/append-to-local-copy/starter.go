package main

import "fmt"

type event struct{ user, action string }

// group раскладывает события по пользователям.
func group(events []event) map[string][]string {
	groups := map[string][]string{}
	for _, e := range events {
		list := groups[e.user]
		list = append(list, e.action)
	}
	return groups
}

func main() {
	events := []event{{"аня", "вход"}, {"боря", "вход"}, {"аня", "покупка"}}
	fmt.Println(group(events))
}
