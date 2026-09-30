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
	out := make(map[string]any, len(base)+len(override))
	for k, v := range base {
		out[k] = deepCopy(v) // даже нетронутые ключи — копия: maps.Clone тут мало
	}
	for k, ov := range override {
		if ov == nil {
			delete(out, k)
			continue
		}
		bm, bok := out[k].(map[string]any)
		om, ook := ov.(map[string]any)
		if bok && ook {
			// bm — уже наша копия, её можно отдавать дальше как base.
			out[k] = Merge(bm, om)
			continue
		}
		out[k] = deepCopy(ov)
	}
	return out
}

// deepCopy копирует контейнеры рекурсивно, скаляры возвращает как есть.
func deepCopy(v any) any {
	switch x := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(x))
		for k, e := range x {
			m[k] = deepCopy(e)
		}
		return m
	case []any:
		s := make([]any, len(x))
		for i, e := range x {
			s[i] = deepCopy(e)
		}
		return s
	default:
		return v
	}
}
