package sakerhet

import (
	"context"
	"fmt"

	abstractedcontainers "github.com/averageflow/sakerhet/pkg/abstracted_containers"
	"github.com/google/uuid"
)

type PostgreSQLIntegrationTestParams struct {
	User     string
	Password string
	DB       string
}

type PostgreSQLIntegrationTester struct {
	TestContext context.Context
	User        string
	Password    string
	DB          string
}

func NewPostgreSQLIntegrationTester(ctx context.Context, p *PostgreSQLIntegrationTestParams) *PostgreSQLIntegrationTester {
	newTester := &PostgreSQLIntegrationTester{}

	if p.Password == "" {
		newTester.Password = fmt.Sprintf("password-%s", uuid.NewString())
	} else {
		newTester.Password = p.Password
	}

	if p.User == "" {
		newTester.User = fmt.Sprintf("user-%s", uuid.NewString())
	} else {
		newTester.User = p.User
	}

	if p.DB == "" {
		newTester.DB = fmt.Sprintf("db-%s", uuid.NewString())
	} else {
		newTester.DB = p.DB
	}

	if ctx == nil {
		newTester.TestContext = context.TODO()
	} else {
		newTester.TestContext = ctx
	}

	return newTester
}

func (p *PostgreSQLIntegrationTester) PrintSomet() {
	fmt.Println(p.DB)
}

func (g *PostgreSQLIntegrationTester) ContainerStart() (*abstractedcontainers.PostgreSQLContainer, error) {
	postgreSQLC, err := abstractedcontainers.SetupPostgreSQL(g.TestContext)
	if err != nil {
		return nil, err
	}

	fmt.Printf("GCP Pub/Sub container started, accessible at: %s\n", postgreSQLC.URI)

	return postgreSQLC, nil
}
