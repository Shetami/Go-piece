package main

// Primes возвращает все простые числа от 2 до limit включительно,
// по возрастанию. Решето Эратосфена — без деления и взятия остатка.
func Primes(limit int) []int {
	if limit < 2 {
		return nil
	}
	composite := make([]bool, limit+1)
	var primes []int
	for n := 2; n <= limit; n++ {
		if composite[n] {
			continue
		}
		primes = append(primes, n)
		// Меньшие кратные уже вычеркнуты меньшими простыми.
		for m := n * n; m <= limit; m += n {
			composite[m] = true
		}
	}
	return primes
}
