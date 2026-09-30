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

func (c *ctxReader) Read(p []byte) (int, error) {
	// ваш код
	return c.r.Read(p)
}
