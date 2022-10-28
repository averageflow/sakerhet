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
	t.Parallel()

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

	initialSchema := []string{
		`
		CREATE TABLE accounts (
				user_id serial PRIMARY KEY,
				username VARCHAR ( 50 ) UNIQUE NOT NULL,
				password VARCHAR ( 50 ) NOT NULL,
				email VARCHAR ( 255 ) UNIQUE NOT NULL,
				age INTEGER NOT NULL,
	      created_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
		`,
	}

	if err := abstractedcontainers.InitPostgreSQLSchema(ctx, dbpool, initialSchema); err != nil {
		t.Fatal(err)
	}

	insertQuery := `INSERT INTO accounts (username, password, email, age) VALUES ($1, $2, $3, $4);`
	seedData := [][]any{
		{"myUser", "myPassword", "myEmail", 25},
	}
	if err := abstractedcontainers.InitPostgreSQLDataInTable(ctx, dbpool, insertQuery, seedData); err != nil {
		t.Fatal(err)
	}
}
