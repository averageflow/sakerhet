package abstractedcontainers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgreSQLContainer struct {
	testcontainers.Container
	URI                     string
	PostgreSQLPort          nat.Port
	PostgreSQLDB            string
	PostgreSQLConnectionURL string
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
		Container:               postgreSQLC,
		URI:                     uri,
		PostgreSQLPort:          postgreSQLPort,
		PostgreSQLDB:            db,
		PostgreSQLConnectionURL: uri,
	}, nil
}

func InitPostgreSQLSchema(ctx context.Context, db *pgxpool.Pool, schema []string) error {
	query := strings.Join(schema, ";\n")

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

func TruncatePostgreSQLTable(ctx context.Context, db *pgxpool.Pool, tables []string) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err != nil {
		return err
	}

	if _, err := tx.Conn().Exec(ctx, fmt.Sprintf(`TRUNCATE TABLE %s;`, strings.Join(tables, ", "))); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func SeedPostgreSQLData(ctx context.Context, db *pgxpool.Pool, query string, data [][]any) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err != nil {
		log.Println(err.Error())
		return err
	}

	for _, v := range data {
		if _, err := tx.Conn().Exec(ctx, query, v...); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
