package main

import (
	"errors"
	"fmt"
)

type Response struct{ Body string }

func (r *Response) Close() { fmt.Println("закрыли", r.Body) }

func fetch(url string) (*Response, error) {
	if url == "" {
		return nil, errors.New("пустой адрес")
	}
	return &Response{Body: url}, nil
}

func get(url string) error {
	resp, err := fetch(url)
	defer resp.Close()
	if err != nil {
		return err
	}
	fmt.Println("прочитали", resp.Body)
	return nil
}

func main() {
	fmt.Println(get("go.dev"))
	fmt.Println(get(""))
}
