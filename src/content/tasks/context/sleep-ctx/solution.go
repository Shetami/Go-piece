package main

import (
	"context"
	"time"
)

// Sleep ждёт d. Если ctx отменят раньше — возвращается сразу с ctx.Err().
func Sleep(ctx context.Context, d time.Duration) error {
	// NewTimer + Stop, а не time.After: таймер остановится сразу,
	// если выйдем по отмене.
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
