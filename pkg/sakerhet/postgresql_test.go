package sakerhet_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	abstractedcontainers "github.com/averageflow/sakerhet/pkg/abstracted_containers"
	"github.com/averageflow/sakerhet/pkg/sakerhet"
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

// before each test
func (suite *PostgreSQLTestSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*30))
	suite.TestContext = ctx
	suite.TestContextCancel = cancel

	suite.IntegrationTester = sakerhet.NewIntegrationTest(sakerhet.IntegrationTesterParams{
		TestContext: ctx,
		PostgreSQL:  &sakerhet.PostgreSQLIntegrationTestParams{},
	})

	postgreSQLC, err := suite.IntegrationTester.PostgreSQLIntegrationTester.ContainerStart()
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.PostgreSQLContainer = postgreSQLC

	dbpool, err := pgxpool.New(suite.TestContext, suite.PostgreSQLContainer.PostgreSQLConnectionURL)
	if err != nil {
		suite.T().Fatal(fmt.Errorf("Unable to create connection pool: %v\n", err))
	}

	suite.DBPool = dbpool
}

func (suite *PostgreSQLTestSuite) TearDownSuite() {
	suite.DBPool.Close()
	_ = suite.PostgreSQLContainer.Terminate(suite.TestContext)
}

func TestPostgreSQLTestSuite(t *testing.T) {
	suite.Run(t, new(PostgreSQLTestSuite))
}

func (suite *PostgreSQLTestSuite) TestHighLevelIntegrationTestPostgreSQL() {
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

	type account struct {
		userId    int
		username  string
		email     string
		age       int
		createdOn int
	}

	situation := PostgreSQLIntegrationTestSituation{
		InitialSchema: []string{
			`
		CREATE TABLE accounts (
				user_id serial PRIMARY KEY,
				username VARCHAR ( 50 ) UNIQUE NOT NULL,
				email VARCHAR ( 255 ) UNIQUE NOT NULL,
				age INTEGER NOT NULL,
	      created_on INTEGER NOT NULL DEFAULT extract(epoch from now())
    );
		`,
		},
		Seeds: []PostgreSQLIntegrationTestSeed{
			{
				InsertQuery: `INSERT INTO accounts (username, email, age, created_on) VALUES ($1, $2, $3, $4);`,
				InsertValues: [][]any{
					{"myUser", "myEmail", 25, 1234567},
				},
			},
		},
		Expects: []PostgreSQLIntegrationTestExpectation{
			{
				GetQuery: `SELECT user_id, username, email, age, created_on FROM accounts;`,
				ExpectedValues: []any{
					account{userId: 1, username: "myUser", email: "myEmail", age: 25, createdOn: 1234567},
				},
			},
		},
	}

	if err := abstractedcontainers.InitPostgreSQLSchema(suite.TestContext, suite.DBPool, situation.InitialSchema); err != nil {
		suite.T().Fatal(err)
	}

	for _, v := range situation.Seeds {
		if err := abstractedcontainers.InitPostgreSQLDataInTable(suite.TestContext, suite.DBPool, v.InsertQuery, v.InsertValues); err != nil {
			suite.T().Fatal(err)
		}
	}

	for _, v := range situation.Expects {
		rows, err := suite.DBPool.Query(suite.TestContext, v.GetQuery)
		if err != nil {
			suite.T().Fatal(err)
		}

		defer rows.Close()

		var result []account

		for rows.Next() {
			var acc account

			if err := rows.Scan(&acc.userId, &acc.username, &acc.email, &acc.age, &acc.createdOn); err != nil {
				suite.T().Fatal(err)
			}

			result = append(result, acc)
		}

		var got []any

		for _, vv := range result {
			got = append(got, vv)
		}

		if !reflect.DeepEqual(got, v.ExpectedValues) {
			suite.T().Fatal(fmt.Errorf(
				"received data is different than expected:\n received %+v\n expected %+v\n",
				got,
				v.ExpectedValues,
			))
		}
	}
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
	if err := abstractedcontainers.InitPostgreSQLDataInTable(ctx, dbpool, insertQuery, seedData); err != nil {
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
