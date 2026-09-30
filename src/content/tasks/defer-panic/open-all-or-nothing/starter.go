package main

import "io"

// Opener описывает, как открыть один ресурс.
type Opener struct {
	Name string
	Open func() (io.Closer, error)
}

// OpenAll открывает ресурсы по порядку: всё или ничего.
//
// Успех: возвращает открытые ресурсы (в порядке openers) и closeAll.
// closeAll закрывает их в обратном порядке, вызывает Close у всех, даже если
// какие-то вернули ошибку, и возвращает errors.Join ошибок в виде
// "close <Name>: <ошибка>". Повторный closeAll ничего не закрывает и
// возвращает nil.
//
// Ошибка Open у k-го ресурса: уже открытые закрываются в обратном порядке,
// OpenAll возвращает nil, nil и errors.Join("open <Name>: <ошибка>", ошибки
// закрытия в том же формате).
//
// Паника в Open: уже открытые закрываются в обратном порядке, а паника
// летит дальше с тем же значением.
func OpenAll(openers []Opener) (res []io.Closer, closeAll func() error, err error) {
	// ваш код
	return nil, nil, nil
}
