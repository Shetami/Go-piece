package main

import "iter"

// Runs группирует подряд идущие элементы seq с одинаковым ключом и отдаёт
// пары (ключ, группа). Одинаковые ключи, разделённые другими, — разные группы.
//   - Каждая группа — отдельный слайс: потребитель может хранить их все.
//   - Ленивая: из seq читается не больше, чем нужно для отданных групп
//     (чтобы понять, что группа кончилась, можно прочитать один элемент следующей).
//   - Досрочный выход из range останавливает и чтение seq.
//   - Результат можно обходить повторно — каждый обход начинается заново.
func Runs[T any, K comparable](seq iter.Seq[T], key func(T) K) iter.Seq2[K, []T] {
	return func(yield func(K, []T) bool) {
		// Состояние — внутри возвращаемой функции: своё у каждого обхода.
		var (
			cur   K
			group []T
		)
		stopped := false
		seq(func(v T) bool {
			k := key(v)
			// Признак «группа есть» — непустой group, а не cur != нулевому
			// ключу: ключ "" или 0 — вполне законный.
			if len(group) > 0 && k != cur {
				if !yield(cur, group) {
					stopped = true
					return false
				}
				group = nil // новый слайс, а не group[:0]: старый уже у потребителя
			}
			cur = k
			group = append(group, v)
			return true
		})
		if !stopped && len(group) > 0 {
			yield(cur, group)
		}
	}
}
