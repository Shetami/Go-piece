package main

import (
	"runtime"
	"sync"
)

// Sampler прореживает логи, как LOG_EVERY_N в glog/klog: для каждого места
// вызова Allow пропускает 1-й, (n+1)-й, (2n+1)-й … вызовы и отклоняет
// остальные. Место вызова — конкретный вызов Allow в исходном коде: два
// вызова в разных строках одной функции, и даже два вызова в одной
// строке — разные места со своими счётчиками. Повторные вызовы из одного
// места (в цикле, из разных горутин) делят один счётчик.
// При n <= 1 Allow всегда возвращает true. Безопасен для конкурентного использования.
type Sampler struct {
	n      int
	mu     sync.Mutex
	counts map[uintptr]int // PC инструкции вызова → сколько раз вызывали
}

func NewSampler(n int) *Sampler {
	return &Sampler{n: n, counts: make(map[uintptr]int)}
}

func (s *Sampler) Allow() bool {
	if s.n <= 1 {
		return true
	}
	// PC адреса возврата в вызывающую функцию: уникален для каждого
	// вызова в коде. Имя функции или file:line различают хуже:
	// два вызова в одной строке дали бы один ключ.
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return true
	}
	s.mu.Lock()
	c := s.counts[pc]
	s.counts[pc] = c + 1
	s.mu.Unlock()
	return c%s.n == 0
}
