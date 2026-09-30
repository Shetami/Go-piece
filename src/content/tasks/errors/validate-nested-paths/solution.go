package main

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrRequired  = errors.New("обязательно")
	ErrFormat    = errors.New("неверный формат")
	ErrRange     = errors.New("вне диапазона")
	ErrDuplicate = errors.New("дубликат")
)

// FieldError — проблема в одном поле. Path — путь от корня проверяемого
// значения: "ID", "Shipping.City", "Items[2].Qty".
type FieldError struct {
	Path string
	Err  error
}

func (e *FieldError) Error() string { return e.Path + ": " + e.Err.Error() }
func (e *FieldError) Unwrap() error { return e.Err }

type Address struct{ City, Zip string }

type Item struct {
	SKU string
	Qty int
}

type Order struct {
	ID       string
	Shipping *Address
	Items    []Item
}

// ValidateAddress и ValidateItem уже написаны и возвращают пути
// относительно своего значения: "City", "Qty".
func ValidateAddress(a Address) error {
	var errs []error
	if a.City == "" {
		errs = append(errs, &FieldError{"City", ErrRequired})
	}
	if len(a.Zip) != 6 || strings.Trim(a.Zip, "0123456789") != "" {
		errs = append(errs, &FieldError{"Zip", ErrFormat})
	}
	return errors.Join(errs...)
}

func ValidateItem(it Item) error {
	var errs []error
	if it.SKU == "" {
		errs = append(errs, &FieldError{"SKU", ErrRequired})
	}
	if it.Qty < 1 || it.Qty > 100 {
		errs = append(errs, &FieldError{"Qty", ErrRange})
	}
	return errors.Join(errs...)
}

// Prefix возвращает копию дерева ошибки err, где путь каждого *FieldError
// (в том числе внутри errors.Join) получил префикс: "City" → "Shipping.City",
// а для префикса "Items[2]" — "Items[2].Qty". Исходные *FieldError НЕ
// меняются. Звенья-не-FieldError переносятся как есть. nil → nil.
func Prefix(prefix string, err error) error {
	switch e := err.(type) {
	case nil:
		return nil
	case *FieldError:
		// Новое значение, а не e.Path = …: исходную ошибку мог вернуть
		// кто-то ещё (кэш, общий сентинел), и мутация испортила бы её.
		return &FieldError{Path: prefix + "." + e.Path, Err: e.Err}
	case interface{ Unwrap() []error }:
		var out []error
		for _, x := range e.Unwrap() {
			out = append(out, Prefix(prefix, x))
		}
		return errors.Join(out...)
	}
	return err
}

// Paths возвращает пути ВСЕХ *FieldError в дереве err в порядке обхода
// (errors.As нашёл бы только первый).
func Paths(err error) []string {
	var out []string
	var walk func(error)
	walk = func(e error) {
		switch x := e.(type) {
		case nil:
		case *FieldError:
			out = append(out, x.Path)
		case interface{ Unwrap() []error }:
			for _, c := range x.Unwrap() {
				walk(c)
			}
		case interface{ Unwrap() error }:
			walk(x.Unwrap())
		}
	}
	walk(err)
	return out
}

// ValidateOrder проверяет заказ и возвращает nil или errors.Join всех
// *FieldError с полными путями, в порядке: ID, Shipping, Items по порядку.
//   - ID пустой                  → "ID": ErrRequired
//   - Shipping == nil            → "Shipping": ErrRequired;
//     иначе ошибки ValidateAddress с префиксом "Shipping"
//   - нет ни одного Items        → "Items": ErrRequired
//   - каждый Items[i]            → ошибки ValidateItem с префиксом "Items[i]"
//   - непустой SKU уже встречался в Items[j] → "Items[i].SKU" с ошибкой,
//     которая оборачивает ErrDuplicate и имеет текст "дубликат: как Items[j]"
//     (j — первое вхождение); идёт после остальных ошибок этого элемента
func ValidateOrder(o Order) error {
	var errs []error
	if o.ID == "" {
		errs = append(errs, &FieldError{"ID", ErrRequired})
	}
	if o.Shipping == nil {
		errs = append(errs, &FieldError{"Shipping", ErrRequired})
	} else {
		errs = append(errs, Prefix("Shipping", ValidateAddress(*o.Shipping)))
	}
	if len(o.Items) == 0 {
		errs = append(errs, &FieldError{"Items", ErrRequired})
	}
	seen := make(map[string]int)
	for i, it := range o.Items {
		p := fmt.Sprintf("Items[%d]", i)
		errs = append(errs, Prefix(p, ValidateItem(it)))
		if it.SKU == "" {
			continue // пустой SKU уже отмечен как обязательный
		}
		if j, ok := seen[it.SKU]; ok {
			errs = append(errs, &FieldError{p + ".SKU", fmt.Errorf("%w: как Items[%d]", ErrDuplicate, j)})
		} else {
			seen[it.SKU] = i
		}
	}
	// Join пропускает nil от валидных частей и схлопывается в nil целиком.
	return errors.Join(errs...)
}
