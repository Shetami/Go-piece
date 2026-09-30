package main

import (
	"context"
	"io"
)

// NewReader оборачивает r так, что Read прерывается отменой ctx.
//   - ctx отменён до Read — r.Read не вызывается, ошибка context.Cause(ctx);
//   - ctx отменили, пока r.Read висит, — Read сразу возвращает (0, Cause),
//     не дожидаясь r; после этого все Read возвращают ту же ошибку;
//   - иначе — ровно то, что вернул r.Read (включая n > 0 вместе с io.EOF).
//
// После возврата из Read обёртка никогда не пишет в переданный p.
func NewReader(ctx context.Context, r io.Reader) io.Reader {
	return &ctxReader{ctx: ctx, r: r}
}

type ctxReader struct {
	ctx context.Context
	r   io.Reader
}

type readResult struct {
	n   int
	err error
}

func (c *ctxReader) Read(p []byte) (int, error) {
	if c.ctx.Err() != nil {
		return 0, context.Cause(c.ctx)
	}
	// Нижний Read пишет в свой буфер: если нас прервут, он может проснуться
	// позже, а p к тому времени вызывающий уже переиспользует.
	buf := make([]byte, len(p))
	done := make(chan readResult, 1) // буфер: прерванная горутина не повиснет на отправке
	go func() {
		n, err := c.r.Read(buf)
		done <- readResult{n, err}
	}()
	select {
	case res := <-done:
		copy(p, buf[:res.n])
		return res.n, res.err
	case <-c.ctx.Done():
		return 0, context.Cause(c.ctx)
	}
}
