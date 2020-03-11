package main

import (
	"context"
	"os"
	"strings"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"
)

type DBWrapper struct {
	Conn *pg.DB
}

type dbLogger struct{}

func (d dbLogger) BeforeQuery(ctx context.Context, q *pg.QueryEvent) (context.Context, error) {
	log := loggerFromContext(ctx)

	query, err := q.FormattedQuery()
	if err != nil {
		log = log.WithError(err)
	}
	log.Info(query)

	return ctx, nil
}

func (d dbLogger) AfterQuery(ctx context.Context, q *pg.QueryEvent) error {
	return nil
}

func NewDBServer(ctx context.Context, config *pg.Options) (*DBWrapper, error) {
	if strings.HasPrefix(config.Password, "$") {
		config.Password = os.Getenv(strings.TrimPrefix(config.Password, "$"))
	}

	db := pg.Connect(config)
	db.AddQueryHook(dbLogger{})

	_, err := db.ExecContext(ctx, "SELECT 1")
	if err != nil {
		return nil, errors.Wrap(err, "cannot ping Postgres")
	}

	return &DBWrapper{
		Conn: db,
	}, nil
}

func (s *DBWrapper) Finalize(ctx context.Context) {
	var err error
	log := loggerFromContext(ctx).WithField("action", "finalize")
	err = s.Conn.Close()
	if err != nil {
		log.WithError(err).Error("error on closing Postgres")
	}
	log.Info("completed")
}
