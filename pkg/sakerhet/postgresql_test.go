package sakerhet_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	abstractedcontainers "github.com/averageflow/sakerhet/pkg/abstracted_containers"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHighLevelIntegrationTestPostgreSQL(t *testing.T) {
	// t.Parallel()

	// given
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
}

func TestLowLevelIntegrationTestPostgreSQL(t *testing.T) {
	// t.Parallel()

	// given
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*30))
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
				email VARCHAR ( 255 ) UNIQUE NOT NULL,
				age INTEGER NOT NULL,
	      created_on INTEGER NOT NULL DEFAULT extract(epoch from now())
    );
		`,
	}

	if err := abstractedcontainers.InitPostgreSQLSchema(ctx, dbpool, initialSchema); err != nil {
		t.Fatal(err)
	}

	insertQuery := `INSERT INTO accounts (username, email, age, created_on) VALUES ($1, $2, $3, $4);`
	seedData := [][]any{
		{"myUser", "myEmail", 25, 1234567},
	}
	if err := abstractedcontainers.InitPostgreSQLDataInTable(ctx, dbpool, insertQuery, seedData); err != nil {
		t.Fatal(err)
	}

	type account struct {
		userId    int
		username  string
		email     string
		age       int
		createdOn int
	}

	// when
	getQuery := `SELECT user_id, username, email, age, created_on FROM accounts;`

	rows, err := dbpool.Query(ctx, getQuery)
	if err != nil {
		t.Fatal(err)
	}

	defer rows.Close()

	var result []account

	for rows.Next() {
		var acc account

		if err := rows.Scan(&acc.userId, &acc.username, &acc.email, &acc.age, &acc.createdOn); err != nil {
			t.Fatal(err)
		}

		result = append(result, acc)
	}

	// then
	expected := []account{
		{userId: 1, username: "myUser", email: "myEmail", age: 25, createdOn: 1234567},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatal(fmt.Errorf(
			"received data is different than expected:\n received %v\n expected %v\n",
			result,
			expected,
		))
	}
}
