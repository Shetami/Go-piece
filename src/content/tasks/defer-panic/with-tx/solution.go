package main

type Tx interface {
	Commit() error
	Rollback() error
}

// WithTx выполняет fn в транзакции:
//   - fn вернула ошибку — Rollback и вернуть ошибку fn;
//   - fn запаниковала — Rollback и паника дальше, с тем же значением;
//   - fn успешна — Commit и вернуть ошибку Commit (если есть).
func WithTx(tx Tx, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// Откатываем и отпускаем панику дальше: решать, что с ней делать,
			// не нам. Ошибку отката здесь некуда вернуть.
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
