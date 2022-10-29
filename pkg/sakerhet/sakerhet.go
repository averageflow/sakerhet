package sakerhet

import "context"

type IntegrationTestParams struct {
	TestContext context.Context
	GCPPubSub   *GCPPubSubIntegrationTestParams
	PostgreSQL  *PostgreSQLIntegrationTestParams
}

type IntegrationTest struct {
	GCPPubSubIntegrationTester  *GCPPubSubIntegrationTester
	PostgreSQLIntegrationTester *PostgreSQLIntegrationTester
}

func NewIntegrationTest(userInput IntegrationTestParams) IntegrationTest {
	var newTest IntegrationTest

	if userInput.GCPPubSub != nil {
		newTest.GCPPubSubIntegrationTester = NewGCPPubSubIntegrationTester(userInput.TestContext, userInput.GCPPubSub)
	}

	if userInput.PostgreSQL != nil {
		newTest.PostgreSQLIntegrationTester = NewPostgreSQLIntegrationTester(userInput.TestContext, userInput.PostgreSQL)
	}

	return newTest
}
