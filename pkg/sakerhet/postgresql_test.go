package sakerhet_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	abstractedcontainers "github.com/averageflow/sakerhet/pkg/abstracted_containers"
)

func TestLowLevelIntegrationTestPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*150))
	defer cancel()

	postgreSQLC, err := abstractedcontainers.SetupPostgreSQL(ctx)
	if err != nil {
		t.Error(err)
	}

	// clean up the container after the test is complete
	defer func() {
		_ = postgreSQLC.Terminate(ctx)
	}()

	fmt.Printf("New container started, accessible at: %s\n", postgreSQLC.URI)

	conn, err := pgx.Connect(context.Background(), postgreSQLC.PostgreSQLConnectionURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())
}
