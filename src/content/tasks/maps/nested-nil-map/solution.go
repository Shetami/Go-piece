package main

import "fmt"

func main() {
	// Кто → к какому ресурсу → какие права.
	acl := map[string]map[string]string{}

	grants := [][3]string{
		{"аня", "отчёты", "чтение"},
		{"аня", "деньги", "запись"},
		{"боря", "отчёты", "запись"},
	}
	for _, g := range grants {
		user, res, perm := g[0], g[1], g[2]
		// Внутренней мапы для нового пользователя ещё нет — создаём при первом обращении.
		if acl[user] == nil {
			acl[user] = map[string]string{}
		}
		acl[user][res] = perm
	}

	fmt.Println(acl)
}
