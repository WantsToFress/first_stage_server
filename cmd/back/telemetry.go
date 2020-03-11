package main

import (
	"context"

	"github.com/sirupsen/logrus"
)

type loggerContextKey struct{}

func contextWithLogger(ctx context.Context, log *logrus.Entry) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, log)
}

func loggerFromContext(ctx context.Context) *logrus.Entry {
	log, ok := ctx.Value(loggerContextKey{}).(*logrus.Entry)
	if !ok {
		return logrus.NewEntry(logrus.New())
	}
	return log
}
