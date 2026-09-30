package main

// Version — разобранная семантическая версия. Build в сравнении не участвует.
type Version struct {
	Major, Minor, Patch int
	Pre                 []string // идентификаторы pre-release: "rc.1" → ["rc", "1"]
	Build               string
}

// ParseVersion разбирает версию по SemVer 2.0.0: MAJOR.MINOR.PATCH,
// за ним необязательный pre-release после '-' и build после '+'.
// Допускается ведущая 'v' ("v1.2.3").
//
//   - MAJOR, MINOR, PATCH — неотрицательные целые без ведущих нулей ("01" — ошибка);
//   - pre-release — идентификаторы через точку, каждый непустой, из
//     [0-9A-Za-z-]; числовой идентификатор без ведущих нулей;
//   - build — идентификаторы через точку, непустые, из [0-9A-Za-z-].
//
// Текст ошибки содержит исходную строку в кавычках, как её печатает %q.
func ParseVersion(s string) (Version, error) {
	// ваш код
	return Version{}, nil
}

// Compare возвращает -1, 0 или 1 по правилам приоритета SemVer.
func (v Version) Compare(w Version) int {
	// ваш код
	return 0
}

// SortVersions возвращает новый слайс версий, отсортированный по
// возрастанию приоритета SemVer. Исходный слайс не меняется. Версии с
// равным приоритетом (отличаются только build) сохраняют исходный порядок.
// Если хоть одна строка некорректна, возвращает nil и ошибку, собранную
// через errors.Join из ошибок всех некорректных строк (в исходном порядке).
func SortVersions(vs []string) ([]string, error) {
	// ваш код
	return vs, nil
}
