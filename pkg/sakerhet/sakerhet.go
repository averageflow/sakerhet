package sakerhet

type SakerhetBuilder struct {
	GCPPubSub  *GCPPubSubIntegrationTestParams
	PostgreSQL *PostgreSQLIntegrationTestParams
}

type Sakerhet struct {
	GCPPubSubIntegrationTester  *GCPPubSubIntegrationTester
	PostgreSQLIntegrationTester *PostgreSQLIntegrationTester
}

// Sakerhet Integration test smart constructor
func NewSakerhetIntegrationTest(sakerhetBuilder SakerhetBuilder) Sakerhet {
	var newTest Sakerhet

	if sakerhetBuilder.GCPPubSub != nil {
		newTest.GCPPubSubIntegrationTester = NewGCPPubSubIntegrationTester(sakerhetBuilder.GCPPubSub)
	}

	if sakerhetBuilder.PostgreSQL != nil {
		newTest.PostgreSQLIntegrationTester = NewPostgreSQLIntegrationTester(sakerhetBuilder.PostgreSQL)
	}

	return newTest
}
