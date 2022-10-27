package sakerhet

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	abstractedcontainers "github.com/averageflow/sakerhet/pkg/abstracted_containers"
)

type IntegrationTest struct {
	GCPPubSubData *IntegrationTestGCPPubSubData
	TestContext   context.Context
	T             *testing.T
	internals     *integrationTestInternals
}

type IntegrationTestGCPPubSubData struct {
	ProjectID      string
	TopicID        string
	SubscriptionID string
}

type integrationTestInternals struct {
	gcpPubSubContainer *abstractedcontainers.GCPPubSubContainer
}

//
// TODO: func NewIntegrationTest | func NewIntegrationTestGCPPubSubData -> smart constructors
// ONLY DO WHEN API IS MORE STABLE
//
// func NewIntegrationTest(userInput IntegrationTest) IntegrationTest {

// }

func initInternalsIfNil(i *IntegrationTest) {
	if i.internals == nil {
		i.internals = &integrationTestInternals{}
	}
}

func (i *IntegrationTest) GCPPubSubContainerStart() (*abstractedcontainers.GCPPubSubContainer, error) {
	topicToSubMap := map[string][]string{i.GCPPubSubData.TopicID: {i.GCPPubSubData.SubscriptionID}}

	pubSubC, err := abstractedcontainers.SetupGCPPubsub(i.TestContext, i.GCPPubSubData.ProjectID, topicToSubMap)
	if err != nil {
		return nil, err
	}

	fmt.Printf("New container started, accessible at: %s\n", pubSubC.URI)

	// required so that all Pub/Sub calls go to docker container, and not GCP
	os.Setenv("PUBSUB_EMULATOR_HOST", pubSubC.URI)

	initInternalsIfNil(i)

	i.internals.gcpPubSubContainer = pubSubC

	return pubSubC, nil
}

func (i *IntegrationTest) GCPPubSubContainsWantedMessages(timeToTimeout time.Duration, expectedData [][]byte) error {
	client, err := pubsub.NewClient(i.TestContext, i.GCPPubSubData.ProjectID)
	if err != nil {
		return err
	}

	defer client.Close()

	if err := abstractedcontainers.CheckGCPMessageInSub(
		i.TestContext,
		client,
		i.GCPPubSubData.SubscriptionID,
		expectedData,
		timeToTimeout,
	); err != nil {
		return err
	}

	return nil
}

func (i *IntegrationTest) GCPPubSubPublishData(wantedData []byte) error {
	client, err := pubsub.NewClient(i.TestContext, i.GCPPubSubData.ProjectID)
	if err != nil {
		return err
	}

	defer client.Close()

	topic, err := abstractedcontainers.GetOrCreateGCPTopic(i.TestContext, client, i.GCPPubSubData.TopicID)
	if err != nil {
		return err
	}

	if err := abstractedcontainers.PublishToGCPTopic(i.TestContext, client, topic, wantedData); err != nil {
		return err
	}

	return nil
}
