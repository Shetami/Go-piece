package main

type Tx interface {
	Commit() error
	Rollback() error
}

// WithTx выполняет fn в транзакции:
//   - fn вернула ошибку — Rollback и вернуть ошибку fn;
//   - fn запаниковала — Rollback и паника дальше, с тем же значением;
//   - fn успешна — Commit и вернуть ошибку Commit (если есть).
func WithTx(tx Tx, fn func() error) error {
	// ваш код
	return fn()
}
