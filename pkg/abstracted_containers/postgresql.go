package abstractedcontainers

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgreSQLContainer struct {
	testcontainers.Container
	Host          string
	MappedPort    string
	DB            string
	ConnectionURL string
}

func SetupPostgreSQL(ctx context.Context, user, pass, db string) (*PostgreSQLContainer, error) {
	postgreSQLPort, err := nat.NewPort("tcp", "5432")
	if err != nil {
		return nil, err
	}

	req := testcontainers.ContainerRequest{
		Image:        "postgres:14.5",
		ExposedPorts: []string{fmt.Sprintf("%s/%s", postgreSQLPort.Port(), postgreSQLPort.Proto())},
		WaitingFor:   wait.ForListeningPort(postgreSQLPort),
		Env: map[string]string{
			"POSTGRES_PASSWORD": pass,
			"POSTGRES_USER":     user,
			"POSTGRES_DB":       db,
		},
		AutoRemove: true,
	}

	postgreSQLC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := postgreSQLC.MappedPort(ctx, postgreSQLPort)
	if err != nil {
		return nil, err
	}

	hostIP, err := postgreSQLC.Host(ctx)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, hostIP, mappedPort.Port(), db)

	return &PostgreSQLContainer{
		Container:     postgreSQLC,
		Host:          hostIP,
		MappedPort:    mappedPort.Port(),
		DB:            db,
		ConnectionURL: uri,
	}, nil
}
