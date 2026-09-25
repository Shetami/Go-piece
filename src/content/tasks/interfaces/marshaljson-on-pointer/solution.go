package main

import (
	"encoding/json"
	"fmt"
)

// Password никогда не должен попадать в JSON в открытом виде.
type Password string

// Метод на значении: так он есть и у Password, и у *Password,
// и json вызовет его, как бы ни передали структуру.
func (p Password) MarshalJSON() ([]byte, error) {
	return []byte(`"***"`), nil
}

type User struct {
	Login    string
	Password Password
}

func main() {
	u := User{Login: "admin", Password: "hunter2"}

	out, err := json.Marshal(u)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}
