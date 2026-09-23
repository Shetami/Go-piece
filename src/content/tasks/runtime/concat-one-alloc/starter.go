package main

// Concat склеивает строки через sep за одно выделение памяти.
func Concat(parts []string, sep string) string {
	// ваш код
	s := ""
	for i, p := range parts {
		if i > 0 {
			s += sep
		}
		s += p
	}
	return s
}
