package sakerhet

type IntegrationTesterParams struct {
	GCPPubSub  *GCPPubSubIntegrationTestParams
	PostgreSQL *PostgreSQLIntegrationTestParams
}

type IntegrationTester struct {
	GCPPubSubIntegrationTester  *GCPPubSubIntegrationTester
	PostgreSQLIntegrationTester *PostgreSQLIntegrationTester
}

func NewIntegrationTest(userInput IntegrationTesterParams) IntegrationTester {
	var newTest IntegrationTester

	if userInput.GCPPubSub != nil {
		newTest.GCPPubSubIntegrationTester = NewGCPPubSubIntegrationTester(userInput.GCPPubSub)
	}

	if userInput.PostgreSQL != nil {
		newTest.PostgreSQLIntegrationTester = NewPostgreSQLIntegrationTester(userInput.PostgreSQL)
	}

	return newTest
}
