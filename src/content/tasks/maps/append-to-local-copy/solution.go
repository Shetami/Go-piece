package main

import "fmt"

type event struct{ user, action string }

// group раскладывает события по пользователям.
func group(events []event) map[string][]string {
	groups := map[string][]string{}
	for _, e := range events {
		// append возвращает новый заголовок слайса — его надо положить обратно в мапу.
		groups[e.user] = append(groups[e.user], e.action)
	}
	return groups
}

func main() {
	events := []event{{"аня", "вход"}, {"боря", "вход"}, {"аня", "покупка"}}
	fmt.Println(group(events))
}
