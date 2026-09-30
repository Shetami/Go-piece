package main

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrMissing — в vars нет значения для переменной шаблона.
	ErrMissing = errors.New("нет значения")
	// ErrSyntax — шаблон синтаксически кривой.
	ErrSyntax = errors.New("ошибка синтаксиса шаблона")
)

// Render подставляет в tpl значения из vars: "{{name}}" → vars["name"].
//
//   - пробелы внутри скобок допустимы: "{{ name }}" — то же, что "{{name}}";
//   - "\{{" выводит буквальные "{{" без подстановки; других экранирований
//     нет, одиночная '\' выводится как есть;
//   - одиночные '{', '}' и "}}" вне подстановки — обычный текст;
//   - подставленное значение не разбирается повторно: "{{x}}" внутри
//     значения остаётся в результате как есть;
//   - незакрытая "{{" или пустое имя ("{{ }}") — ошибка, оборачивающая
//     ErrSyntax;
//   - если нескольких переменных нет в vars, возвращается errors.Join из
//     ошибок по каждой отсутствующей — по одной на имя, в порядке первого
//     появления; каждая оборачивает ErrMissing и содержит имя.
//
// При любой ошибке результат — "".
func Render(tpl string, vars map[string]string) (string, error) {
	var b strings.Builder
	var missing []error
	seen := map[string]bool{}
	for {
		i := strings.Index(tpl, "{{")
		if i < 0 {
			b.WriteString(tpl)
			break
		}
		// "\{{" — экранированные скобки: пишем "{{" и идём дальше.
		if i > 0 && tpl[i-1] == '\\' {
			b.WriteString(tpl[:i-1])
			b.WriteString("{{")
			tpl = tpl[i+2:]
			continue
		}
		b.WriteString(tpl[:i])
		rest := tpl[i+2:]
		name, after, ok := strings.Cut(rest, "}}")
		if !ok {
			return "", fmt.Errorf("%w: незакрытая {{", ErrSyntax)
		}
		name = strings.TrimSpace(name)
		if name == "" {
			return "", fmt.Errorf("%w: пустое имя переменной", ErrSyntax)
		}
		if v, ok := vars[name]; ok {
			b.WriteString(v) // значение пишем как есть, повторно не разбираем
		} else if !seen[name] {
			seen[name] = true
			missing = append(missing, fmt.Errorf("%w: %s", ErrMissing, name))
		}
		tpl = after
	}
	if len(missing) > 0 {
		return "", errors.Join(missing...)
	}
	return b.String(), nil
}
