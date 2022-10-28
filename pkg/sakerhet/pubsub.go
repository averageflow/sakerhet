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
	TestContext    context.Context
	ProjectID      string
	TopicID        string
	SubscriptionID string
	PubSubURI      string
}

func NewGCPPubSubIntegrationTester(ctx context.Context, g *GCPPubSubIntegrationTestParams) *GCPPubSubIntegrationTester {
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

	if ctx == nil {
		newTester.TestContext = context.TODO()
	} else {
		newTester.TestContext = ctx
	}

	return newTester
}

func (g *GCPPubSubIntegrationTester) ContainerStart() (*abstractedcontainers.GCPPubSubContainer, error) {
	topicToSubMap := map[string][]string{g.TopicID: {g.SubscriptionID}}

	pubSubC, err := abstractedcontainers.SetupGCPPubsub(g.TestContext, g.ProjectID, topicToSubMap)
	if err != nil {
		return nil, err
	}

	fmt.Printf("New container started, accessible at: %s\n", pubSubC.URI)

	// required so that all Pub/Sub calls go to docker container, and not GCP
	// os.Setenv("PUBSUB_EMULATOR_HOST", pubSubC.URI)
	g.PubSubURI = pubSubC.URI

	return pubSubC, nil
}

func (g *GCPPubSubIntegrationTester) createClient() (*pubsub.Client, error) {
	conn, err := grpc.Dial(g.PubSubURI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc.Dial: %v", err)
	}

	o := []option.ClientOption{
		option.WithGRPCConn(conn),
		option.WithTelemetryDisabled(),
	}

	client, err := pubsub.NewClientWithConfig(g.TestContext, g.ProjectID, nil, o...)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (g *GCPPubSubIntegrationTester) ContainsWantedMessages(timeToTimeout time.Duration, expectedData [][]byte) error {
	// client, err := pubsub.NewClient(g.TestContext, g.ProjectID)
	// if err != nil {
	// 	return err
	// }
	//
	// defer client.Close()
	// conn, err := grpc.Dial(g.PubSubURI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	return fmt.Errorf("grpc.Dial: %v", err)
	// }

	// o := []option.ClientOption{
	// 	option.WithGRPCConn(conn),
	// 	option.WithTelemetryDisabled(),
	// }

	// client, err := pubsub.NewClientWithConfig(g.TestContext, g.ProjectID, nil, o...)
	// if err != nil {
	// 	return err
	// }

	client, err := g.createClient()
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
	client, err := g.createClient()
	if err != nil {
		return err
	}

	defer client.Close()
	// client, err := pubsub.NewClient(g.TestContext, g.ProjectID)
	// if err != nil {
	// 	return err
	// }

	// defer client.Close()

	topic, err := abstractedcontainers.GetOrCreateGCPTopic(g.TestContext, client, g.TopicID)
	if err != nil {
		return err
	}

	if err := abstractedcontainers.PublishToGCPTopic(g.TestContext, client, topic, wantedData); err != nil {
		return err
	}

	return nil
}
