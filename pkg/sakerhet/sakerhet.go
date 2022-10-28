package sakerhet

type IntegrationTest struct {
	GCPPubSubIntegrationTester *GCPPubSubIntegrationTester
}

func NewIntegrationTest(userInput IntegrationTest) IntegrationTest {
	var result IntegrationTest

	if userInput.GCPPubSubIntegrationTester != nil {
		result.GCPPubSubIntegrationTester = userInput.GCPPubSubIntegrationTester
	}

	return result
}
