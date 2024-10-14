package test

import (
	"context"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type Cleanup func()
type Port int

func SetupTestContainersDb() (Port, error, Cleanup) {
	pgCt, err := postgres.Run(context.Background(),
		"postgres:16-alpine",
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithDatabase("postgres"),
		postgres.WithInitScripts("../../init.sql"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return 0, err, nil
	}
	port, err := pgCt.MappedPort(context.Background(), "5432")
	if err != nil {
		return 0, err, nil
	}
	return Port(port.Int()), nil, func() {
		pgCt.Terminate(context.Background())
	}
}
