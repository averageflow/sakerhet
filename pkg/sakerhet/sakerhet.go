package sakerhet

import "context"

type IntegrationTestParams struct {
	TestContext context.Context
	GCPPubSub   *GCPPubSubIntegrationTestParams
}

type IntegrationTest struct {
	GCPPubSubIntegrationTester *GCPPubSubIntegrationTester
}

func NewIntegrationTest(userInput IntegrationTestParams) IntegrationTest {
	var newTest IntegrationTest

	if userInput.GCPPubSub != nil {
		newTest.GCPPubSubIntegrationTester = NewGCPPubSubIntegrationTester(userInput.TestContext, userInput.GCPPubSub)
	}

	return newTest
}
