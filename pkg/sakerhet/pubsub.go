package sakerhet

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
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

func (g *GCPPubSubIntegrationTester) ContainsWantedMessages(ctx context.Context, expectedData [][]byte) error {
	return g.ContainsWantedMessagesInDuration(ctx, expectedData, 1500*time.Millisecond)
}

func (g *GCPPubSubIntegrationTester) ContainsWantedMessagesInDuration(ctx context.Context, expectedData [][]byte, timeToTimeout time.Duration) error {
	client, err := g.CreateClient(ctx)
	if err != nil {
		return err
	}

	defer client.Close()

	if err := ExpectGCPMessagesInSub(
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

func (g *GCPPubSubIntegrationTester) ReadMessages(ctx context.Context) ([][]byte, error) {
	return g.ReadMessagesInDuration(ctx, 1500*time.Millisecond)
}

func (g *GCPPubSubIntegrationTester) ReadMessagesInDuration(ctx context.Context, timeToTimeout time.Duration) ([][]byte, error) {
	client, err := g.CreateClient(ctx)
	if err != nil {
		return nil, err
	}

	defer client.Close()

	messages, err := ReadGCPMessagesInSub(
		ctx,
		client,
		g.SubscriptionID,
		timeToTimeout,
	)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (g *GCPPubSubIntegrationTester) PublishData(ctx context.Context, wantedData []byte) error {
	client, err := g.CreateClient(ctx)
	if err != nil {
		return err
	}

	defer client.Close()

	topic, err := GetOrCreateGCPTopic(ctx, client, g.TopicID)
	if err != nil {
		return err
	}

	if err := PublishToGCPTopic(ctx, client, topic, wantedData); err != nil {
		return err
	}

	return nil
}

func toReadableSliceOfStrings(raw [][]byte) []string {
	result := make([]string, len(raw))

	for i := range raw {
		result[i] = string(raw[i])
	}

	return result
}

// Receive messages for a given duration, which simplifies testing.
func ExpectGCPMessagesInSub(ctx context.Context, client *pubsub.Client, subscriptionID string, expectedData [][]byte, timeToWait time.Duration) error {
	sub := client.Subscription(subscriptionID)

	ctx, cancel := context.WithTimeout(ctx, timeToWait)
	defer cancel()

	var receivedData [][]byte

	mu := &sync.Mutex{}

	err := sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		mu.Lock()
		receivedData = append(receivedData, msg.Data)
		mu.Unlock()

		msg.Ack()
	})
	if err != nil {
		return fmt.Errorf("sub.Receive: %v", err)
	}

	if !UnorderedEqual(expectedData, receivedData) {
		return fmt.Errorf(
			"received data is different than expected:\n received %v\n expected %v\n",
			toReadableSliceOfStrings(receivedData),
			toReadableSliceOfStrings(expectedData),
		)
	}

	return nil
}

// Receive messages for a given duration, which simplifies testing.
func ReadGCPMessagesInSub(ctx context.Context, client *pubsub.Client, subscriptionID string, timeToWait time.Duration) ([][]byte, error) {
	sub := client.Subscription(subscriptionID)

	ctx, cancel := context.WithTimeout(ctx, timeToWait)
	defer cancel()

	var receivedData [][]byte

	mu := &sync.Mutex{}

	err := sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		mu.Lock()
		receivedData = append(receivedData, msg.Data)
		mu.Unlock()

		msg.Ack()
	})
	if err != nil {
		return nil, fmt.Errorf("sub.Receive: %v", err)
	}

	return receivedData, nil
}

func GetOrCreateGCPTopic(ctx context.Context, client *pubsub.Client, topicID string) (*pubsub.Topic, error) {
	topic := client.Topic(topicID)

	ok, err := topic.Exists(ctx)
	if err != nil {
		return nil, err
	}

	if !ok {
		if _, err = client.CreateTopic(ctx, topicID); err != nil {
			return nil, err
		}
	}

	return topic, nil
}

func PublishToGCPTopic(ctx context.Context, client *pubsub.Client, topic *pubsub.Topic, payload []byte) error {
	var wg sync.WaitGroup
	var totalErrors uint64

	result := topic.Publish(ctx, &pubsub.Message{
		Data: payload,
	})

	wg.Add(1)

	go func(res *pubsub.PublishResult) {
		defer wg.Done()
		// The Get method blocks until a server-generated ID or
		// an error is returned for the published message.
		_, err := res.Get(ctx)
		if err != nil {
			// Error handling code can be added here.
			atomic.AddUint64(&totalErrors, 1)
			return
		}
	}(result)

	wg.Wait()

	if totalErrors > 0 {
		return fmt.Errorf("%d messages did not publish successfully", totalErrors)
	}

	return nil
}
