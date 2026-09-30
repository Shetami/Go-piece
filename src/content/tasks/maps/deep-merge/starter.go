package main

// Merge накладывает override поверх base и возвращает новый документ
// (JSON-подобный: map[string]any, []any и скаляры).
//
//   - Если с обеих сторон map[string]any — они сливаются рекурсивно.
//   - Иначе значение из override заменяет значение из base; слайсы
//     заменяются целиком, а не склеиваются.
//   - Значение nil в override удаляет ключ (на любой глубине).
//   - base и override не меняются, и результат не делит с ними ни одной
//     вложенной map[string]any или []any: правка результата на любой глубине
//     не видна во входах, и наоборот.
//   - nil вместо base или override — то же, что пустая мапа.
func Merge(base, override map[string]any) map[string]any {
	// ваш код
	return nil
}
