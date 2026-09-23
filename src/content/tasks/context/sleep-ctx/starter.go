package main

import (
	"context"
	"time"
)

// Sleep ждёт d. Если ctx отменят раньше — возвращается сразу с ctx.Err().
func Sleep(ctx context.Context, d time.Duration) error {
	// ваш код
	time.Sleep(d)
	return nil
}
