package main

import (
	"errors"
	"fmt"
	"io"
)

// ErrClosed — Read после Close.
var ErrClosed = errors.New("concat: read after close")

type concat struct {
	parts  []io.ReadCloser // ещё не закрытые части, текущая — первая
	idx    int             // номер текущей части в исходном списке
	closed bool
}

// Concat читает части подряд, как io.MultiReader, но заботится о закрытии:
//   - часть закрывается сразу, как только вернула io.EOF, — не дожидаясь
//     общего Close (держать открытыми сотни файлов незачем);
//   - если это закрытие вернуло ошибку, Read возвращает
//     fmt.Errorf("part %d: %w", i, err) (i — индекс части) и уже прочитанные n;
//   - io.EOF части — не конец: Read с n > 0 и EOF части возвращает n, nil;
//     общий io.EOF — только когда кончились все части;
//   - другая ошибка части возвращается как есть, часть не закрывается
//     (её закроет общий Close);
//   - Close закрывает все ещё не закрытые части (и недочитанные, и не
//     начатые), по порядку, ошибки — errors.Join; повторный Close — nil;
//   - Read после Close — 0, ErrClosed.
func Concat(parts ...io.ReadCloser) io.ReadCloser {
	return &concat{parts: parts}
}

func (c *concat) Read(p []byte) (int, error) {
	if c.closed {
		return 0, ErrClosed
	}
	for len(c.parts) > 0 {
		n, err := c.parts[0].Read(p)
		if err == io.EOF {
			cur := c.parts[0]
			c.parts[0] = nil
			c.parts = c.parts[1:] // из списка — до Close: второй раз не закроем
			i := c.idx
			c.idx++
			if cerr := cur.Close(); cerr != nil {
				return n, fmt.Errorf("part %d: %w", i, cerr)
			}
			if n > 0 {
				return n, nil // EOF части — не EOF всего потока
			}
			continue
		}
		if n > 0 || err != nil {
			return n, err
		}
		// 0, nil — пробуем ту же часть ещё раз
	}
	return 0, io.EOF
}

func (c *concat) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	var errs []error
	for _, rc := range c.parts {
		if err := rc.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	c.parts = nil
	return errors.Join(errs...)
}
