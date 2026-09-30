package main

import (
	"errors"
	"io"
	"strings"
)

// ErrClosed — Read после Close.
var ErrClosed = errors.New("concat: read after close")

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
	// ваш код
	return io.NopCloser(strings.NewReader(""))
}
