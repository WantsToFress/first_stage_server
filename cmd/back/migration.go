package main

import (
	"github.com/go-pg/pg/v9"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/prometheus/common/log"
	"time"
)

type MigrationConfig struct {
	Path string `yaml:"path"`
}

func Migrate(dbConfig *pg.Options, migraionConfig MigrationConfig) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err, _ = r.(error)
		}
	}()

	connectionString := "postgres://" + dbConfig.User + ":" + dbConfig.Password + "@" + dbConfig.Addr + "/" + dbConfig.Database + "?sslmode=disable"
	var m *migrate.Migrate

	for i := 0; i < 60; i++ {
		m, err = migrate.New(migraionConfig.Path, connectionString)
		if err != nil {
			log.Error(err)
			time.Sleep(time.Second * 1)
			continue
		}
		break
	}
	if err != nil {
		return err
	}

	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			err = sourceErr
		}
		if dbErr != nil {
			err = dbErr
		}
	}()

	if err := m.Up(); err != nil {
		return err
	}

	return nil
}
