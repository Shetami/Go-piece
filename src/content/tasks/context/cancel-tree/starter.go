package main

import (
	"context"
	"fmt"
)

func main() {
	parent, cancelParent := context.WithCancel(context.Background())
	child, cancelChild := context.WithCancel(parent)

	cancelChild()
	fmt.Println(parent.Err(), child.Err())

	child2, cancel2 := context.WithCancel(parent)
	defer cancel2()

	cancelParent()
	fmt.Println(parent.Err(), child2.Err())

	<-child2.Done()
	fmt.Println("done закрыт")
}
