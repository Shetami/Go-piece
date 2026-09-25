package main

import (
	"fmt"
	"slices"
	"sync"
)

type Request struct {
	ID   int
	User string
}

func handle(r *Request, wg *sync.WaitGroup, out chan<- string) {
	defer wg.Done()
	out <- fmt.Sprintf("%d:%s", r.ID, r.User)
}

func main() {
	users := []string{"аня", "боря", "вика"}
	out := make(chan string, len(users))

	var wg sync.WaitGroup
	for i, u := range users {
		// Своя переменная на каждую итерацию: горутины не делят один Request.
		req := Request{ID: i + 1, User: u}
		wg.Add(1)
		go handle(&req, &wg, out)
	}
	wg.Wait()
	close(out)

	var got []string
	for s := range out {
		got = append(got, s)
	}
	slices.Sort(got)
	fmt.Println(got)
}
