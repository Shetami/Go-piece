package main

import (
	"context"
	"errors"
	"fmt"
)

func main() {
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(errors.New("клиент ушёл"))
	cancel(errors.New("второй раз"))

	fmt.Println(ctx.Err())
	fmt.Println(context.Cause(ctx))

	ctx2, cancel2 := context.WithCancel(context.Background())
	cancel2()
	fmt.Println(context.Cause(ctx2))

	detached := context.WithoutCancel(ctx)
	fmt.Println(detached.Err(), detached.Done() == nil)
}
