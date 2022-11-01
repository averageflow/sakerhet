package sakerhet_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	abstractedcontainers "github.com/averageflow/sakerhet/pkg/abstracted_containers"
	"github.com/averageflow/sakerhet/pkg/sakerhet"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

type PostgreSQLTestSuite struct {
	suite.Suite
	TestContext         context.Context
	TestContextCancel   context.CancelFunc
	PostgreSQLContainer *abstractedcontainers.PostgreSQLContainer
	IntegrationTester   sakerhet.IntegrationTester
	DBPool              *pgxpool.Pool
}

// Before suite starts
func (suite *PostgreSQLTestSuite) SetupSuite() {
	ctx := context.Background()

	suite.IntegrationTester = sakerhet.NewIntegrationTest(sakerhet.IntegrationTesterParams{
		PostgreSQL: &sakerhet.PostgreSQLIntegrationTestParams{},
	})

	// create PostgreSQL container
	postgreSQLC, err := suite.IntegrationTester.PostgreSQLIntegrationTester.ContainerStart(ctx)
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.PostgreSQLContainer = postgreSQLC

	// create DB pool that will be reused across tests
	dbpool, err := pgxpool.New(ctx, suite.PostgreSQLContainer.PostgreSQLConnectionURL)
	if err != nil {
		suite.T().Fatal(fmt.Errorf("Unable to create connection pool: %v\n", err))
	}

	suite.DBPool = dbpool

	// setup schema that will be reused across tests
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
	integrationTestTimeout := int64(60)

	if givenTimeout := os.Getenv("SAKERHET_INTEGRATION_TEST_TIMEOUT"); givenTimeout != "" {
		if x, err := strconv.ParseInt(givenTimeout, 10, 64); err == nil {
			integrationTestTimeout = x
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(integrationTestTimeout)*time.Second)
	suite.TestContext = ctx
	suite.TestContextCancel = cancel
}

// After each test
func (suite *PostgreSQLTestSuite) TearDownTest() {
	ctx := context.Background()
	if err := suite.IntegrationTester.PostgreSQLIntegrationTester.TruncateTable(ctx, suite.DBPool, []string{"accounts"}); err != nil {
		suite.T().Fatal(err)
	}
}

// After suite finishes
func (suite *PostgreSQLTestSuite) TearDownSuite() {
	suite.DBPool.Close()
	_ = suite.PostgreSQLContainer.Terminate(suite.TestContext)
}

func TestPostgreSQLTestSuite(t *testing.T) {
	if os.Getenv("SAKERHET_RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration tests! Set variable SAKERHET_RUN_INTEGRATION_TESTS to run them!")
	} else {
		suite.Run(t, new(PostgreSQLTestSuite))
	}
}

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

	log.Printf("%+v", suite.DBPool)
	log.Printf("%+v", suite.TestContext)

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

func TestLowLevelIntegrationTestPostgreSQL(t *testing.T) {
	if os.Getenv("SAKERHET_RUN_INTEGRATION_TESTS") == "" {
		t.Skip()
	}

	// given
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*60))
	defer cancel()

	password := fmt.Sprintf("password-%s", uuid.NewString())
	user := fmt.Sprintf("user-%s", uuid.NewString())
	db := fmt.Sprintf("db-%s", uuid.NewString())

	postgreSQLC, err := abstractedcontainers.SetupPostgreSQL(ctx, user, password, db)
	if err != nil {
		t.Fatal(err)
	}

	// clean up the container after the test is complete
	defer func() {
		_ = postgreSQLC.Terminate(ctx)
	}()

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

	// when
	if err := abstractedcontainers.SeedPostgreSQLData(ctx, dbpool, insertQuery, seedData); err != nil {
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