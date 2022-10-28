package sakerhet_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	abstractedcontainers "github.com/averageflow/sakerhet/pkg/abstracted_containers"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestLowLevelIntegrationTestPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*15))
	defer cancel()

	postgreSQLC, err := abstractedcontainers.SetupPostgreSQL(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// clean up the container after the test is complete
	defer func() {
		_ = postgreSQLC.Terminate(ctx)
	}()

	fmt.Printf("New container started, accessible at: %s\n", postgreSQLC.URI)

	dbpool, err := pgxpool.New(ctx, postgreSQLC.PostgreSQLConnectionURL)
	if err != nil {
		t.Fatal(fmt.Errorf("Unable to create connection pool: %v\n", err))
	}

	defer dbpool.Close()

	// var greeting string
	// err = dbpool.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
	// 	os.Exit(1)
	// }
	// conn, err := pgx.Connect(context.Background(), postgreSQLC.PostgreSQLConnectionURL)
	// if err != nil {
	// 	t.Fatal(fmt.Errorf("Unable to connect to database: %v\n", err))
	// }tx
	//
	// defer conn.Close(context.Background())

	initialSchema := []string{
		`
		CREATE TABLE accounts (
				user_id serial PRIMARY KEY,
				username VARCHAR ( 50 ) UNIQUE NOT NULL,
				password VARCHAR ( 50 ) NOT NULL,
				email VARCHAR ( 255 ) UNIQUE NOT NULL,
	      created_on TIMESTAMP NOT NULL,
        last_login TIMESTAMP 
    );
		`,
	}

	abstractedcontainers.InitPostgreSQLSchema(ctx, dbpool, initialSchema)
}
