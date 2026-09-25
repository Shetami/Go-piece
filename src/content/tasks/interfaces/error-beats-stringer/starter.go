package main

import "fmt"

type status struct{ code int }

func (s status) String() string { return fmt.Sprintf("статус %d", s.code) }
func (s status) Error() string  { return fmt.Sprintf("ошибка %d", s.code) }

type temp float64

func (t temp) String() string { return fmt.Sprintf("%.1f°C", float64(t)) }

func main() {
	s := status{404}
	fmt.Println(s)
	fmt.Printf("%v | %s | %d\n", s, s, s)

	t := temp(21.5)
	fmt.Println(t)
	fmt.Printf("%.2f\n", t)
}
