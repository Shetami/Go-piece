package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	<-ctx.Done()
	fmt.Println(ctx.Err())
	fmt.Println(errors.Is(ctx.Err(), context.DeadlineExceeded))

	parent, cancelParent := context.WithCancel(context.Background())
	cancelParent()
	child, cancelChild := context.WithTimeout(parent, time.Hour)
	defer cancelChild()
	fmt.Println(child.Err())

	early, cancelEarly := context.WithTimeout(context.Background(), time.Hour)
	cancelEarly()
	fmt.Println(early.Err())
	_, ok := early.Deadline()
	fmt.Println(ok)
}
