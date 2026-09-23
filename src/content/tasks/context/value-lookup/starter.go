package main

import (
	"context"
	"fmt"
)

type key string

func main() {
	ctx := context.WithValue(context.Background(), key("user"), "Аня")
	child := context.WithValue(ctx, key("user"), "Боря")

	fmt.Println(ctx.Value(key("user")), child.Value(key("user")))
	fmt.Println(child.Value("user"))
	fmt.Println(child.Value(key("role")))
}
