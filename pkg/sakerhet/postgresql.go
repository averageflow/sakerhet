package sakerhet

import (
	"context"
	"fmt"

	abstractedcontainers "github.com/averageflow/sakerhet/pkg/abstracted_containers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

type PostgreSQLIntegrationTestSeed struct {
	InsertQuery  string
	InsertValues [][]any
}

type PostgreSQLIntegrationTestExpectation struct {
	GetQuery       string
	ExpectedValues []any
}

type PostgreSQLIntegrationTestSituation struct {
	InitialSchema []string
	Seeds         []PostgreSQLIntegrationTestSeed
	Expects       []PostgreSQLIntegrationTestExpectation
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

func (g *PostgreSQLIntegrationTester) ContainerStart() (*abstractedcontainers.PostgreSQLContainer, error) {
	postgreSQLC, err := abstractedcontainers.SetupPostgreSQL(g.TestContext)
	if err != nil {
		return nil, err
	}

	return postgreSQLC, nil
}

func (g *PostgreSQLIntegrationTester) InitSchema(dbPool *pgxpool.Pool, initialSchema []string) error {
	if err := abstractedcontainers.InitPostgreSQLSchema(g.TestContext, dbPool, initialSchema); err != nil {
		return err
	}

	return nil
}

func (p *PostgreSQLIntegrationTester) SeedData(dbPool *pgxpool.Pool, seeds []PostgreSQLIntegrationTestSeed) error {
	for _, v := range seeds {
		if err := abstractedcontainers.InitPostgreSQLDataInTable(p.TestContext, dbPool, v.InsertQuery, v.InsertValues); err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgreSQLIntegrationTester) CheckContainsExpectedData(resultSet []any, expected []any) error {
	if !abstractedcontainers.UnorderedEqual(resultSet, expected) {
		return fmt.Errorf(
			"received data is different than expected:\n received %+v\n expected %+v\n",
			resultSet,
			expected,
		)
	}

	return nil
}

func (p *PostgreSQLIntegrationTester) FetchData(dbPool *pgxpool.Pool, query string, rowHandler func(rows pgx.Rows) (any, error)) ([]any, error) {
	rows, err := dbPool.Query(p.TestContext, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []any

	for rows.Next() {
		x, err := rowHandler(rows)
		if err != nil {
			return nil, err
		}

		result = append(result, x)
	}

	return result, nil
}
