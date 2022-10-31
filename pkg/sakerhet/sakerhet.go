package sakerhet

import "context"

type IntegrationTesterParams struct {
	TestContext context.Context
	GCPPubSub   *GCPPubSubIntegrationTestParams
	PostgreSQL  *PostgreSQLIntegrationTestParams
}

type IntegrationTester struct {
	GCPPubSubIntegrationTester  *GCPPubSubIntegrationTester
	PostgreSQLIntegrationTester *PostgreSQLIntegrationTester
}

func NewIntegrationTest(userInput IntegrationTesterParams) IntegrationTester {
	var newTest IntegrationTester

	if userInput.GCPPubSub != nil {
		newTest.GCPPubSubIntegrationTester = NewGCPPubSubIntegrationTester(userInput.TestContext, userInput.GCPPubSub)
	}

	if userInput.PostgreSQL != nil {
		newTest.PostgreSQLIntegrationTester = NewPostgreSQLIntegrationTester(userInput.TestContext, userInput.PostgreSQL)
	}

	return newTest
}
