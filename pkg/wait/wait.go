package wait

import (
	"context"
	"time"
)

func Seconds(s time.Duration) {
	time.Sleep(s * time.Second)
}

func CtxSeconds(s time.Duration) (context.Context, time.Duration) {
	return context.Background(), s * time.Second
}

func CtxMinutes(m time.Duration) (context.Context, time.Duration) {
	return context.Background(), m * time.Minute
}
