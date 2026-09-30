package main

import (
	"math/bits"
	"sync"
)

// Размеры классов пула: степени двойки от MinClass до MaxClass включительно.
const (
	MinClass = 64
	MaxClass = 64 << 10
)

const (
	minShift   = 6  // 1<<6 == MinClass
	maxShift   = 16 // 1<<16 == MaxClass
	numClasses = maxShift - minShift + 1
)

// BytePool раздаёт []byte из корзин по степеням двойки — как пулы буферов
// в сетевых библиотеках. Нулевое значение готово к работе, пул безопасен
// для конкурентного использования.
type BytePool struct {
	classes [numClasses]sync.Pool
}

// Get возвращает слайс длины n (n >= 0). Если n <= MaxClass, вместимость
// слайса равна ровно размеру класса — наименьшей степени двойки,
// которая >= max(n, MinClass). Иначе — обычный make([]byte, n).
// Содержимое слайса не гарантируется.
func (p *BytePool) Get(n int) []byte {
	if n > MaxClass {
		return make([]byte, n)
	}
	// Округление вверх: класс, в который n гарантированно влезет.
	shift := minShift
	if n > MinClass {
		shift = bits.Len(uint(n - 1))
	}
	if v := p.classes[shift-minShift].Get(); v != nil {
		// Храним *[]byte: слайс в any копировался бы в кучу на каждом Put.
		return (*v.(*[]byte))[:n]
	}
	return make([]byte, n, 1<<shift)
}

// Put отдаёт слайс обратно. Слайсы с cap < MinClass или cap > MaxClass
// выбрасываются. Слайс с cap не степенью двойки тоже годится в дело.
func (p *BytePool) Put(b []byte) {
	c := cap(b)
	if c < MinClass || c > MaxClass {
		return
	}
	// Округление ВНИЗ: слайс с cap 100 годится только для класса 64.
	// Положи его в класс 128 — и Get(120) сделает b[:120] за пределами cap.
	shift := bits.Len(uint(c)) - 1
	b = b[:0:1<<shift] // вместимость ровно размер класса
	p.classes[shift-minShift].Put(&b)
}
