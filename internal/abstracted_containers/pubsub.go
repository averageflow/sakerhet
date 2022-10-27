package abstractedcontainers

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PubSubContainer struct {
	testcontainers.Container
	URI string
}

func SetupPubsub(ctx context.Context) (*PubSubContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "bigtruedata/gcloud-pubsub-emulator",
		ExposedPorts: []string{"8538/tcp"},
		WaitingFor:   wait.ForExposedPort(),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	ip, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "8538")
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())
	log.Printf("New container started, accessible at: %s", uri)

	return &PubSubContainer{Container: container, URI: uri}, nil
}

func GetOrCreateTopic(uri, topicID string) (*pubsub.Topic, error) {
	ctx := context.Background()

	os.Setenv("PUBSUB_EMULATOR_HOST", uri)

	client, err := pubsub.NewClient(ctx, "test_project")
	if err != nil {
		return nil, err
	}

	defer client.Close()

	topic := client.Topic(topicID)

	// Create the topic if it doesn't exist.
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, err
	}

	if !exists {
		log.Printf("Topic %v doesn't exist - creating it", topicID)

		if _, err = client.CreateTopic(ctx, topicID); err != nil {
			return nil, err
		}
	}

	return topic, nil
}

func SendMessageToPubSub(uri string, topic *pubsub.Topic) error {
	ctx := context.Background()

	os.Setenv("PUBSUB_EMULATOR_HOST", uri)

	msg := &pubsub.Message{
		Data: []byte("This is a test message!"),
	}

	if _, err := topic.Publish(ctx, msg).Get(ctx); err != nil {
		return err
	}

	log.Println("Message published.")
	return nil
}

func CheckMessageReceived(uri string, topic *pubsub.Topic, subscriptionID string) error {
	ctx := context.Background()

	os.Setenv("PUBSUB_EMULATOR_HOST", uri)

	client, err := pubsub.NewClient(ctx, "test_project")
	if err != nil {
		return err
	}

	sub, err := client.CreateSubscription(ctx, subscriptionID, pubsub.SubscriptionConfig{
		Topic: topic,
	})

	cctx, cancel := context.WithCancel(ctx)
	okCh := make(chan string)

	go func() {
		// Use a callback to receive messages via subscription.
		// Receive will block until the context is cancelled, or we get a non-recoverable error
		err = sub.Receive(cctx, func(_ context.Context, m *pubsub.Message) {
			m.Ack()
			okCh <- string(m.Data)
			cancel()
		})
	}()

	select {
	case msg := <-okCh:
		if msg != "This is a test message!" {
			return fmt.Errorf("")
		}
	case <-time.After(300 * time.Millisecond):
		cancel()
		return fmt.Errorf("did not receive message within deadline")
	}

	return nil
}
