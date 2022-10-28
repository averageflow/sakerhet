package abstractedcontainers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type postgreSQLContainer struct {
	testcontainers.Container
	URI                     string
	PostgreSQLPort          nat.Port
	PostgreSQLDB            string
	PostgreSQLConnectionURL string
}

func SetupPostgreSQL(ctx context.Context) (*postgreSQLContainer, error) {
	postgreSQLPort, err := nat.NewPort("tcp", "5432")
	if err != nil {
		return nil, err
	}

	postgreSQLPassword := fmt.Sprintf("password-%s", uuid.NewString())
	postgreSQLUser := fmt.Sprintf("user-%s", uuid.NewString())
	postgreSQLDB := fmt.Sprintf("db-%s", uuid.NewString())

	req := testcontainers.ContainerRequest{
		Image:        "postgres:14.5",
		ExposedPorts: []string{fmt.Sprintf("%s/%s", postgreSQLPort.Port(), postgreSQLPort.Proto())},
		WaitingFor:   wait.NewHostPortStrategy(postgreSQLPort),
		Env: map[string]string{
			"POSTGRES_PASSWORD": postgreSQLPassword,
			"POSTGRES_USER":     postgreSQLUser,
			"POSTGRES_DB":       postgreSQLDB,
		},
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

	// postgres://user:secret@localhost:5432/mydatabasename
	uri := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", postgreSQLUser, postgreSQLPassword, hostIP, mappedPort.Port(), postgreSQLDB)

	return &postgreSQLContainer{
		Container:               postgreSQLC,
		URI:                     uri,
		PostgreSQLPort:          postgreSQLPort,
		PostgreSQLDB:            postgreSQLDB,
		PostgreSQLConnectionURL: uri,
	}, nil
}

func initPostgreSQL(ctx context.Context, db sql.DB) error {
	// Actual SQL for initializing the database should probably live elsewhere
	const query = `CREATE DATABASE projectmanagement;
        CREATE TABLE projectmanagement.task(
            id uuid primary key not null,
            description varchar(255) not null,
            date_due timestamp with time zone,
            date_created timestamp with time zone not null,
            date_updated timestamp with time zone not null);`
	_, err := db.ExecContext(ctx, query)

	return err
}

func truncatePostgreSQL(ctx context.Context, db sql.DB) error {
	const query = `TRUNCATE projectmanagement.task`

	_, err := db.ExecContext(ctx, query)
	return err
}
