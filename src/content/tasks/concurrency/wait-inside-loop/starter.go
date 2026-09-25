package main

import (
	"fmt"
	"sync"
	"time"
)

func download(name string) {
	time.Sleep(time.Second)
}

func main() {
	files := []string{"a.zip", "b.zip", "c.zip"}
	start := time.Now()

	var wg sync.WaitGroup
	for _, f := range files {
		wg.Add(1)
		go func() {
			defer wg.Done()
			download(f)
		}()
		wg.Wait()
	}

	fmt.Println("скачано:", len(files), "за", time.Since(start).Round(time.Second))
}
