package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), -time.Second)
	defer cancel()
	fmt.Println(ctx.Err())
	fmt.Println(errors.Is(ctx.Err(), context.DeadlineExceeded), errors.Is(ctx.Err(), context.Canceled))

	parent, cancelParent := context.WithTimeout(context.Background(), time.Hour)
	defer cancelParent()
	child, cancelChild := context.WithTimeout(parent, 2*time.Hour)
	defer cancelChild()

	d1, _ := parent.Deadline()
	d2, _ := child.Deadline()
	fmt.Println(d1.Equal(d2))

	_, ok := context.Background().Deadline()
	fmt.Println(ok)
}
