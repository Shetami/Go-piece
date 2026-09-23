package main

import "fmt"

type Temp float64

func (t Temp) String() string { return fmt.Sprintf("%.1f°C", float64(t)) }
func (t Temp) Error() string  { return "перегрев" }

func describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "ничего"
	case fmt.Stringer:
		return "stringer " + x.String()
	case error:
		return "error " + x.Error()
	case float64:
		return "float64"
	default:
		return fmt.Sprintf("%T", x)
	}
}

func main() {
	fmt.Println(describe(Temp(36.6)))
	fmt.Println(describe(36.6))
	fmt.Println(describe(nil))
	fmt.Println(describe(error(nil)))
	fmt.Println(Temp(40))
}
