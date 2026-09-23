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
	if err != nil {
		return err
	}
	// Откладываем закрытие только после проверки: при ошибке resp — nil.
	defer resp.Close()
	fmt.Println("прочитали", resp.Body)
	return nil
}

func main() {
	fmt.Println(get("go.dev"))
	fmt.Println(get(""))
}
