package sakerhet

import (
	"context"
	"fmt"
	"strings"

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
	User     string
	Password string
	DB       string
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
	Seeds   []PostgreSQLIntegrationTestSeed
	Expects []PostgreSQLIntegrationTestExpectation
}

func NewPostgreSQLIntegrationTester(p *PostgreSQLIntegrationTestParams) *PostgreSQLIntegrationTester {
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

	return newTester
}

func (g *PostgreSQLIntegrationTester) ContainerStart(ctx context.Context) (*abstractedcontainers.PostgreSQLContainer, error) {
	postgreSQLC, err := abstractedcontainers.SetupPostgreSQL(ctx, g.User, g.Password, g.DB)
	if err != nil {
		return nil, err
	}

	return postgreSQLC, nil
}

func (p *PostgreSQLIntegrationTester) InitSchema(ctx context.Context, dbPool *pgxpool.Pool, initialSchema []string) error {
	if err := InitPostgreSQLSchema(ctx, dbPool, initialSchema); err != nil {
		return err
	}

	return nil
}

func (p *PostgreSQLIntegrationTester) SeedData(ctx context.Context, dbPool *pgxpool.Pool, seeds []PostgreSQLIntegrationTestSeed) error {
	for _, v := range seeds {
		if err := SeedPostgreSQLData(ctx, dbPool, v.InsertQuery, v.InsertValues); err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgreSQLIntegrationTester) CheckContainsExpectedData(resultSet []any, expected []any) error {
	if !UnorderedEqual(resultSet, expected) {
		return fmt.Errorf(
			"received data is different than expected:\n received %+v\n expected %+v\n",
			resultSet,
			expected,
		)
	}

	return nil
}

func (p *PostgreSQLIntegrationTester) TruncateTable(ctx context.Context, dbPool *pgxpool.Pool, tables []string) error {
	return TruncatePostgreSQLTable(ctx, dbPool, tables)
}

func (p *PostgreSQLIntegrationTester) FetchData(ctx context.Context, dbPool *pgxpool.Pool, query string, rowHandler func(rows pgx.Rows) (any, error)) ([]any, error) {
	rows, err := dbPool.Query(ctx, query)
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
		return err
	}

	for _, v := range data {
		if _, err := tx.Conn().Exec(ctx, query, v...); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
