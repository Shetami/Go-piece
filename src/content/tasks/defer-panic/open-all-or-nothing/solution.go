package main

import (
	"errors"
	"fmt"
	"io"
)

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
	var names []string
	// Открытое копим в отдельной переменной, а не в именованном res:
	// return nil, nil, err обнулит res раньше, чем выполнится defer.
	var opened []io.Closer
	success := false
	// Один defer на оба неудачных пути: и на return с ошибкой, и на панику.
	// recover не нужен — паника после отложенных вызовов полетит дальше сама.
	defer func() {
		if !success {
			err = errors.Join(err, closeReverse(names, opened))
		}
	}()

	for _, o := range openers {
		c, oerr := o.Open()
		if oerr != nil {
			return nil, nil, fmt.Errorf("open %s: %w", o.Name, oerr)
		}
		names = append(names, o.Name)
		opened = append(opened, c)
	}
	success = true

	closed := false
	return opened, func() error {
		if closed {
			return nil
		}
		closed = true
		return closeReverse(names, opened)
	}, nil
}

// closeReverse закрывает всё в обратном порядке и собирает ошибки.
func closeReverse(names []string, cs []io.Closer) error {
	var errs []error
	for i := len(cs) - 1; i >= 0; i-- {
		if err := cs[i].Close(); err != nil {
			errs = append(errs, fmt.Errorf("close %s: %w", names[i], err))
		}
	}
	return errors.Join(errs...)
}
