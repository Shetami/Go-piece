package main

import (
	"context"
	"fmt"
)

type userKey struct{}

func process(ctx context.Context) {
	user, ok := ctx.Value(userKey{}).(string)
	if !ok {
		fmt.Println("process: аноним")
		return
	}
	fmt.Println("process:", user)
}

func handle(ctx context.Context, token string) {
	if token != "" {
		user := "user:" + token
		ctx := context.WithValue(ctx, userKey{}, user)
		fmt.Println("авторизован:", ctx.Value(userKey{}))
	}
	process(ctx)
}

func main() {
	handle(context.Background(), "anna")
}
