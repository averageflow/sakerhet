package abstractedcontainers

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
		WaitingFor:   wait.ForListeningPort(postgreSQLPort),
		Env: map[string]string{
			"POSTGRES_PASSWORD": postgreSQLPassword,
			"POSTGRES_USER":     postgreSQLUser,
			"POSTGRES_DB":       postgreSQLDB,
		},
		Name: "postgresql",
	}

	postgreSQLC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            true,
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

func InitPostgreSQLSchema(ctx context.Context, db *pgxpool.Pool, schema []string) error {
	query := strings.Join(schema, ";\n")

	fmt.Printf("DB: %+v", db)
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err != nil {
		return err
	}

	if _, err := tx.Conn().Exec(ctx, query); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func InitPostgreSQLDataInTable(ctx context.Context, db *pgxpool.Pool, query string, data [][]any) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err != nil {
		return err
	}

	for _, v := range data {
		if _, err := tx.Conn().Exec(ctx, query, v...); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// func truncatePostgreSQL(ctx context.Context, db sql.DB) error {
// 	const query = `TRUNCATE projectmanagement.task`
//
// 	_, err := db.ExecContext(ctx, query)
// 	return err
// }
