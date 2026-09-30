package main

import (
	"context"
	"errors"
	"fmt"
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
	if k > len(replicas) {
		return Reply{}, ErrNoQuorum
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		r   Reply
		err error
	}
	results := make(chan result, len(replicas)) // опоздавшие не блокируются
	for _, rep := range replicas {
		go func() {
			r, err := rep(ctx)
			results <- result{r, err}
		}()
	}

	var best Reply
	oks := 0
	var errs []error
	allowedFailures := len(replicas) - k
	for oks < k {
		select {
		case res := <-results:
			if res.err != nil {
				errs = append(errs, res.err)
				if len(errs) > allowedFailures { // кворум уже не собрать
					return Reply{}, fmt.Errorf("%w: %w", ErrNoQuorum, errors.Join(errs...))
				}
				continue
			}
			if oks == 0 || res.r.Version > best.Version {
				best = res.r
			}
			oks++
		case <-ctx.Done():
			return Reply{}, ctx.Err()
		}
	}
	return best, nil
}
