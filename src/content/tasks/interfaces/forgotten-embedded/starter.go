package main

import "fmt"

type Store interface {
	Get(key string) string
}

type mapStore map[string]string

func (m mapStore) Get(key string) string { return m[key] }

// LoggedStore логирует обращения и передаёт их дальше.
type LoggedStore struct {
	Store
}

func (l LoggedStore) Get(key string) string {
	fmt.Println("get", key)
	return l.Store.Get(key)
}

func WithLogging(s Store) Store {
	return LoggedStore{}
}

func main() {
	s := WithLogging(mapStore{"lang": "go"})
	fmt.Println(s.Get("lang"))
}
