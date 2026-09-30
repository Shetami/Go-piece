package main

import (
	"errors"
	"strings"
)

var (
	// ErrTraversal — путь пытается выйти за пределы корня.
	ErrTraversal = errors.New("выход за пределы корня")
	// ErrBadPath — в пути недопустимый символ.
	ErrBadPath = errors.New("недопустимый путь")
)

// SafeJoin склеивает корневую директорию root с путём rel, пришедшим от
// пользователя (из URL файлового сервера, имени файла в архиве и т. п.).
//
//   - root — уже чистый абсолютный путь без '/' в конце: "/srv/files";
//     для root == "/" результат начинается с одного '/';
//   - rel считается относительным от root, даже если начинается с '/';
//   - разделители — и '/', и '\' (Windows-клиенты и хитрые атаки);
//     повторные разделители схлопываются, "." убирается, ".." поднимается
//     на уровень вверх;
//   - если ".." поднимается выше root — ErrTraversal (не обрезать молча!);
//   - "..." и "..a" — обычные имена;
//   - нулевой байт в rel — ErrBadPath;
//   - результат без '/' в конце; пустой rel (или сведённый к пустому) — root.
//
// Функция работает только со строками и файловую систему не трогает.
func SafeJoin(root, rel string) (string, error) {
	if strings.IndexByte(rel, 0) >= 0 {
		return "", ErrBadPath
	}
	// Слайс как стек сегментов: ".." снимает верхний.
	var stack []string
	parts := strings.FieldsFunc(rel, func(r rune) bool { return r == '/' || r == '\\' })
	for _, p := range parts {
		switch p {
		case ".":
		case "..":
			if len(stack) == 0 {
				return "", ErrTraversal
			}
			stack = stack[:len(stack)-1]
		default:
			stack = append(stack, p)
		}
	}
	if len(stack) == 0 {
		return root, nil
	}
	return strings.TrimSuffix(root, "/") + "/" + strings.Join(stack, "/"), nil
}
