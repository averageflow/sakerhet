package sakerhet_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	dummyservices "github.com/averageflow/sakerhet/pkg/dummy_services"
	"github.com/averageflow/sakerhet/pkg/sakerhet"
	"github.com/stretchr/testify/suite"
)

// Test suite demonstrating the use of the HTTP utilities of Säkerhet
type HttpTestSuite struct {
	suite.Suite
	testApi dummyservices.SomeTestAPI
}

// Before suite starts
func (suite *HttpTestSuite) SetupSuite() {
	suite.testApi = dummyservices.NewSomeTestAPI()
	go suite.testApi.Start()
}

// After suite ends
func (suite *HttpTestSuite) TearDownSuite() {
	suite.testApi.Stop()
}

// Start the test suite if we are running integration tests
func TestHttpTestSuite(t *testing.T) {
	sakerhet.SkipIntegrationTestsWhenUnitTesting(t)
	t.Parallel()
	suite.Run(t, new(HttpTestSuite))
}

// High level test on code that performs HTTP requests and expects a given response
func (suite *HttpTestSuite) TestHighLevelHttpCalls() {
	situation := &sakerhet.HttpIntegrationTestSituation{
		Request: &sakerhet.HttpIntegrationTestSituationRequest{
			RequestURL:    "http://localhost:8765/custom-status-response",
			RequestMethod: http.MethodPost,
			RequestHeaders: []sakerhet.HttpHeaderValuePair{
				{Header: "Content-Type", Value: "application/json"},
			},
			RequestBody: []byte(`{"wantedResponseCode": 200}`),
		},
		Expectation: &sakerhet.HttpIntegrationTestSituationExpectation{
			ResponseStatusCode: http.StatusOK,
			ResponseBody:       []byte(fmt.Sprintf(`{"returnedResponseCode": %d}`, http.StatusOK)),
		},
		Timeout: 2 * time.Second,
	}

	if _, err := situation.SituationChecker(); err != nil {
		suite.T().Fatal(err.Error())
	}
}

// Low level test with full control performing HTTP requests and expecting given response
func (suite *HttpTestSuite) TestLowLevelHttpCalls() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	postURL := "http://localhost:8765/custom-status-response"
	body := []byte(`{"wantedResponseCode": 200}`)

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, postURL, bytes.NewBuffer(body))
	if err != nil {
		suite.T().Fatal(err.Error())
	}

	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}

	res, err := client.Do(r)
	if err != nil {
		suite.T().Fatal(err.Error())
	}

	defer res.Body.Close()

	received, err := io.ReadAll(res.Body)
	if err != nil {
		suite.T().Fatal(err.Error())
	}

	expected := fmt.Sprintf(`{"returnedResponseCode": %d}`, http.StatusOK)

	if string(received) != expected {
		suite.T().Fatalf("Unexpected data received! Expected %v, got %v", string(expected), string(received))
	}

	if res.StatusCode != http.StatusOK {
		suite.T().Fatalf("Unexpected status code received! Expected %d, got %d", http.StatusOK, res.StatusCode)
	}
}
