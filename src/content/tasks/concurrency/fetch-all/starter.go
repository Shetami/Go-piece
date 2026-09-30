package main

// FetchAll параллельно загружает все urls через fetch — по горутине на
// каждый уникальный URL. Повторяющиеся URL загружаются один раз.
//
// Возвращает карту url → тело для успешных загрузок (не nil, даже если
// успешных нет) и ошибку: errors.Join всех неудач в порядке первого
// появления URL во входе, каждая обёрнута как fmt.Errorf("%s: %w", url, err).
// Если неудач нет — nil.
func FetchAll(urls []string, fetch func(url string) (string, error)) (map[string]string, error) {
	// ваш код
	return nil, nil
}
