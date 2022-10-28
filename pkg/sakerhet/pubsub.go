package sakerhet

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	abstractedcontainers "github.com/averageflow/sakerhet/pkg/abstracted_containers"
)

type GCPPubSubIntegrationTester struct {
	TestContext    context.Context
	ProjectID      string
	TopicID        string
	SubscriptionID string
}

func (g *GCPPubSubIntegrationTester) ContainerStart() (*abstractedcontainers.GCPPubSubContainer, error) {
	topicToSubMap := map[string][]string{g.TopicID: {g.SubscriptionID}}

	pubSubC, err := abstractedcontainers.SetupGCPPubsub(g.TestContext, g.ProjectID, topicToSubMap)
	if err != nil {
		return nil, err
	}

	fmt.Printf("New container started, accessible at: %s\n", pubSubC.URI)

	// required so that all Pub/Sub calls go to docker container, and not GCP
	os.Setenv("PUBSUB_EMULATOR_HOST", pubSubC.URI)

	return pubSubC, nil
}

func (g *GCPPubSubIntegrationTester) ContainsWantedMessages(timeToTimeout time.Duration, expectedData [][]byte) error {
	client, err := pubsub.NewClient(g.TestContext, g.ProjectID)
	if err != nil {
		return err
	}

	defer client.Close()

	if err := abstractedcontainers.CheckGCPMessageInSub(
		g.TestContext,
		client,
		g.SubscriptionID,
		expectedData,
		timeToTimeout,
	); err != nil {
		return err
	}

	return nil
}

func (g *GCPPubSubIntegrationTester) PublishData(wantedData []byte) error {
	client, err := pubsub.NewClient(g.TestContext, g.ProjectID)
	if err != nil {
		return err
	}

	defer client.Close()

	topic, err := abstractedcontainers.GetOrCreateGCPTopic(g.TestContext, client, g.TopicID)
	if err != nil {
		return err
	}

	if err := abstractedcontainers.PublishToGCPTopic(g.TestContext, client, topic, wantedData); err != nil {
		return err
	}

	return nil
}
