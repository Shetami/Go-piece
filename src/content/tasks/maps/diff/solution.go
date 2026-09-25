package main

import "slices"

// Diff сравнивает два снимка настроек и возвращает отсортированные
// по возрастанию списки ключей: добавленных, удалённых и изменённых.
// Если в какой-то категории ключей нет — пустой слайс или nil, неважно.
func Diff(old, cur map[string]string) (added, removed, changed []string) {
	for k, v := range cur {
		// ok отличает «ключа не было» от «было пустое значение».
		prev, ok := old[k]
		switch {
		case !ok:
			added = append(added, k)
		case prev != v:
			changed = append(changed, k)
		}
	}
	for k := range old {
		if _, ok := cur[k]; !ok {
			removed = append(removed, k)
		}
	}
	// Порядок обхода мапы случаен — без сортировки результат плавал бы.
	slices.Sort(added)
	slices.Sort(removed)
	slices.Sort(changed)
	return added, removed, changed
}
