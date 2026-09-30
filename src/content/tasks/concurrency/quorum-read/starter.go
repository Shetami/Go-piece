package main

import (
	"context"
	"errors"
)

var ErrNoQuorum = errors.New("no quorum")

// Reply — ответ реплики: значение и его версия.
type Reply struct {
	Value   string
	Version int
}

// Replica читает значение с одной реплики.
type Replica func(ctx context.Context) (Reply, error)

// QuorumRead опрашивает все реплики параллельно.
//
//   - Как только k реплик ответили успешно, возвращает ответ с наибольшей
//     Version среди этих k, не дожидаясь остальных.
//   - Как только кворум стал недостижим (упало больше len(replicas)-k),
//     сразу возвращает ошибку, не дожидаясь остальных. Ошибка оборачивает
//     ErrNoQuorum и все ошибки реплик, полученные к этому моменту.
//     k > len(replicas) — сразу ErrNoQuorum без опроса.
//   - При возврате контекст оставшихся реплик отменяется; ни одна горутина
//     не остаётся заблокированной навсегда.
//   - Отмена ctx — ctx.Err().
func QuorumRead(ctx context.Context, replicas []Replica, k int) (Reply, error) {
	// ваш код
	return Reply{}, nil
}
