package utils

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
)

var Log *SentryWrapper

type SentryWrapper struct{}

func InitLogger(dsn string) error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:        dsn,
		EnableLogs: true,
	})
	if err != nil {
		return err
	}
	Log = &SentryWrapper{}
	return nil
}

func (sw *SentryWrapper) Info(ctx context.Context) sentry.LogEntry {
	return sentry.NewLogger(ctx).Info()
}

func (sw *SentryWrapper) Error(ctx context.Context) sentry.LogEntry {
	return sentry.NewLogger(ctx).Error()
}

func (sw *SentryWrapper) Debug(ctx context.Context) sentry.LogEntry {
	return sentry.NewLogger(ctx).Debug()
}

func SyncLogger() {
	sentry.Flush(2 * time.Second)
}
