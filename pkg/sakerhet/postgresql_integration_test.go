package sakerhet_test

import (
	"context"
	"fmt"
	"testing"

	abstractedcontainers "github.com/averageflow/sakerhet/pkg/abstracted_containers"
	"github.com/averageflow/sakerhet/pkg/sakerhet"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

// Test suite demonstrating the use of Säkerhet + testcontainers for PostgreSQL
type PostgreSQLTestSuite struct {
	suite.Suite
	TestContext         context.Context
	TestContextCancel   context.CancelFunc
	PostgreSQLContainer *abstractedcontainers.PostgreSQLContainer
	IntegrationTester   sakerhet.Sakerhet
	DBPool              *pgxpool.Pool
}

// Before suite starts
func (suite *PostgreSQLTestSuite) SetupSuite() {
	ctx := context.Background()

	suite.IntegrationTester = sakerhet.NewSakerhetIntegrationTest(sakerhet.SakerhetBuilder{
		PostgreSQL: &sakerhet.PostgreSQLIntegrationTestParams{},
	})

	// Spin up one PostgreSQL container for all the tests in the suite
	postgreSQLC, err := suite.IntegrationTester.PostgreSQLIntegrationTester.ContainerStart(ctx)
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.PostgreSQLContainer = postgreSQLC

	// Create DB pool that will be reused across tests
	dbpool, err := pgxpool.New(ctx, suite.PostgreSQLContainer.ConnectionURL)
	if err != nil {
		suite.T().Fatal(fmt.Errorf("Unable to create connection pool: %v\n", err))
	}

	suite.DBPool = dbpool

	// Setup schema that will be reused across tests
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

	if err := suite.IntegrationTester.PostgreSQLIntegrationTester.InitSchema(ctx, suite.DBPool, initialSchema); err != nil {
		suite.T().Fatal(err)
	}
}

// Before each test
func (suite *PostgreSQLTestSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), sakerhet.GetIntegrationTestTimeout())
	suite.TestContext = ctx
	suite.TestContextCancel = cancel
}

// After each test
func (suite *PostgreSQLTestSuite) TearDownTest() {
	if err := suite.IntegrationTester.PostgreSQLIntegrationTester.TruncateTable(
		context.Background(),
		suite.DBPool,
		[]string{"accounts"},
	); err != nil {
		suite.T().Fatal(err)
	}
}

// After suite ends
func (suite *PostgreSQLTestSuite) TearDownSuite() {
	suite.DBPool.Close()
	_ = suite.PostgreSQLContainer.Terminate(context.Background())
}

// Start the test suite if we are running integration tests
func TestPostgreSQLTestSuite(t *testing.T) {
	sakerhet.SkipIntegrationTestsWhenUnitTesting(t)
	suite.Run(t, new(PostgreSQLTestSuite))
}

// High level test on code that uses PostgreSQL
func (suite *PostgreSQLTestSuite) TestHighLevelIntegrationTestPostgreSQL() {
	type account struct {
		userId    int
		username  string
		email     string
		age       int
		createdOn int
	}

	situation := sakerhet.PostgreSQLIntegrationTestSituation{
		Seeds: []sakerhet.PostgreSQLIntegrationTestSeed{
			{
				InsertQuery: `INSERT INTO accounts (username, email, age, created_on) VALUES ($1, $2, $3, $4);`,
				InsertValues: [][]any{
					{"myUser", "myEmail", 25, 1234567},
					{"mySecondUser", "mySecondEmail", 50, 999999},
				},
			},
		},
		Expects: []sakerhet.PostgreSQLIntegrationTestExpectation{
			{
				GetQuery: `SELECT user_id, username, email, age, created_on FROM accounts;`,
				ExpectedValues: []any{
					account{userId: 2, username: "mySecondUser", email: "mySecondEmail", age: 50, createdOn: 999999},
					account{userId: 1, username: "myUser", email: "myEmail", age: 25, createdOn: 1234567},
				},
			},
		},
	}

	if err := suite.IntegrationTester.PostgreSQLIntegrationTester.SeedData(suite.TestContext, suite.DBPool, situation.Seeds); err != nil {
		suite.T().Fatal(err)
	}

	rowHandler := func(rows pgx.Rows) (any, error) {
		var acc account

		if err := rows.Scan(&acc.userId, &acc.username, &acc.email, &acc.age, &acc.createdOn); err != nil {
			return nil, err
		}

		return acc, nil
	}

	for _, v := range situation.Expects {
		got, err := suite.IntegrationTester.PostgreSQLIntegrationTester.FetchData(suite.TestContext, suite.DBPool, v.GetQuery, rowHandler)
		if err != nil {
			suite.T().Fatal(err)
		}

		if err := suite.IntegrationTester.PostgreSQLIntegrationTester.CheckContainsExpectedData(got, v.ExpectedValues); err != nil {
			suite.T().Fatal(err)
		}
	}
}

// Low level test with full control on testing code that uses PostgreSQL
func TestLowLevelIntegrationTestPostgreSQL(t *testing.T) {
	sakerhet.SkipIntegrationTestsWhenUnitTesting(t)

	// given
	password := fmt.Sprintf("password-%s", uuid.NewString())
	user := fmt.Sprintf("user-%s", uuid.NewString())
	db := fmt.Sprintf("db-%s", uuid.NewString())

	postgreSQLC, err := abstractedcontainers.SetupPostgreSQL(context.Background(), user, password, db)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), sakerhet.GetIntegrationTestTimeout())
	defer cancel()

	// clean up the container after the test is complete
	defer func() {
		_ = postgreSQLC.Terminate(context.Background())
	}()

	dbpool, err := pgxpool.New(ctx, postgreSQLC.ConnectionURL)
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

	if err := sakerhet.InitPostgreSQLSchema(ctx, dbpool, initialSchema); err != nil {
		t.Fatal(err)
	}

	insertQuery := `INSERT INTO accounts (username, email, age, created_on) VALUES ($1, $2, $3, $4);`
	seedData := [][]any{
		{"myUser", "myEmail", 25, 1234567},
	}

	// when
	if err := sakerhet.SeedPostgreSQLData(ctx, dbpool, insertQuery, seedData); err != nil {
		t.Fatal(err)
	}

	// then
	type account struct {
		userId    int
		username  string
		email     string
		age       int
		createdOn int
	}

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

	expected := []account{
		{userId: 1, username: "myUser", email: "myEmail", age: 25, createdOn: 1234567},
	}

	if !sakerhet.UnorderedEqual(result, expected) {
		t.Fatal(fmt.Errorf(
			"received data is different than expected:\n received %v\n expected %v\n",
			result,
			expected,
		))
	}
}
