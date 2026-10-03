// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/sieve, MIT, © Exercism). Руками не править.
package main

func TestPrimes(t *testing.T) {
	cases := []struct {
		name  string
		limit int
		want  []int
	}{
		{"no primes under two", 1, []int{}},
		{"find first prime", 2, []int{2}},
		{"find primes up to 10", 10, []int{2, 3, 5, 7}},
		{"limit is prime", 13, []int{2, 3, 5, 7, 11, 13}},
		{"find primes up to 1000", 1000, []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97, 101, 103, 107, 109, 113, 127, 131, 137, 139, 149, 151, 157, 163, 167, 173, 179, 181, 191, 193, 197, 199, 211, 223, 227, 229, 233, 239, 241, 251, 257, 263, 269, 271, 277, 281, 283, 293, 307, 311, 313, 317, 331, 337, 347, 349, 353, 359, 367, 373, 379, 383, 389, 397, 401, 409, 419, 421, 431, 433, 439, 443, 449, 457, 461, 463, 467, 479, 487, 491, 499, 503, 509, 521, 523, 541, 547, 557, 563, 569, 571, 577, 587, 593, 599, 601, 607, 613, 617, 619, 631, 641, 643, 647, 653, 659, 661, 673, 677, 683, 691, 701, 709, 719, 727, 733, 739, 743, 751, 757, 761, 769, 773, 787, 797, 809, 811, 821, 823, 827, 829, 839, 853, 857, 859, 863, 877, 881, 883, 887, 907, 911, 919, 929, 937, 941, 947, 953, 967, 971, 977, 983, 991, 997}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Primes(c.limit)
			if !sameResult(got, c.want) {
				t.Fatalf("Primes(%v) = %v, ожидали %v", c.limit, got, c.want)
			}
		})
	}
}

// sameResult — reflect.DeepEqual, который не отличает nil от пустого слайса
// или мапы: для ответа это одно и то же.
func sameResult(got, want any) bool {
	g, w := reflect.ValueOf(got), reflect.ValueOf(want)
	if g.Kind() == w.Kind() && (g.Kind() == reflect.Slice || g.Kind() == reflect.Map) && g.Len() == 0 && w.Len() == 0 {
		return true
	}
	return reflect.DeepEqual(got, want)
}
