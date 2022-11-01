package sakerhet

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	abstractedcontainers "github.com/averageflow/sakerhet/pkg/abstracted_containers"
	"github.com/google/uuid"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GCPPubSubIntegrationTestParams struct {
	ProjectID      string
	TopicID        string
	SubscriptionID string
}

type GCPPubSubIntegrationTester struct {
	ProjectID      string
	TopicID        string
	SubscriptionID string
	PubSubURI      string
}

func NewGCPPubSubIntegrationTester(g *GCPPubSubIntegrationTestParams) *GCPPubSubIntegrationTester {
	newTester := &GCPPubSubIntegrationTester{}

	if g.ProjectID == "" {
		newTester.ProjectID = "test-project-" + uuid.New().String()
	} else {
		newTester.ProjectID = g.ProjectID
	}

	if g.TopicID == "" {
		newTester.TopicID = "test-topic-" + uuid.New().String()
	} else {
		newTester.TopicID = g.TopicID
	}

	if g.SubscriptionID == "" {
		newTester.SubscriptionID = "test-sub-" + uuid.New().String()
	} else {
		newTester.SubscriptionID = g.SubscriptionID
	}

	return newTester
}

func (g *GCPPubSubIntegrationTester) ContainerStart(ctx context.Context) (*abstractedcontainers.GCPPubSubContainer, error) {
	topicToSubMap := map[string][]string{g.TopicID: {g.SubscriptionID}}

	pubSubC, err := abstractedcontainers.SetupGCPPubsub(ctx, g.ProjectID, topicToSubMap)
	if err != nil {
		return nil, err
	}

	g.PubSubURI = pubSubC.URI

	return pubSubC, nil
}

func (g *GCPPubSubIntegrationTester) CreateClient(ctx context.Context) (*pubsub.Client, error) {
	conn, err := grpc.Dial(g.PubSubURI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc.Dial: %v", err)
	}

	o := []option.ClientOption{
		option.WithGRPCConn(conn),
		option.WithTelemetryDisabled(),
	}

	client, err := pubsub.NewClientWithConfig(ctx, g.ProjectID, nil, o...)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (g *GCPPubSubIntegrationTester) ContainsWantedMessages(ctx context.Context, timeToTimeout time.Duration, expectedData [][]byte) error {
	client, err := g.CreateClient(ctx)
	if err != nil {
		return err
	}

	defer client.Close()

	if err := abstractedcontainers.AwaitGCPMessageInSub(
		ctx,
		client,
		g.SubscriptionID,
		expectedData,
		timeToTimeout,
	); err != nil {
		return err
	}

	return nil
}

func (g *GCPPubSubIntegrationTester) PublishData(ctx context.Context, wantedData []byte) error {
	client, err := g.CreateClient(ctx)
	if err != nil {
		return err
	}

	defer client.Close()

	topic, err := abstractedcontainers.GetOrCreateGCPTopic(ctx, client, g.TopicID)
	if err != nil {
		return err
	}

	if err := abstractedcontainers.PublishToGCPTopic(ctx, client, topic, wantedData); err != nil {
		return err
	}

	return nil
}
