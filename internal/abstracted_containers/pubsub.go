package abstractedcontainers

import (
	"context"
	"fmt"
	"log"
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
		WaitingFor:   wait.ForLog("[pubsub] INFO: Server started, listening on 8538"),
		NetworkMode:  "bridge",
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

	uri := fmt.Sprintf("%s:%s", ip, mappedPort.Port())
	// uri := fmt.Sprintf("172.17.0.1:%s", mappedPort.Port())
	log.Printf("New container started, accessible at: %s", uri)

	return &PubSubContainer{Container: container, URI: uri}, nil
}

func GetOrCreateTopic(ctx context.Context, uri, topicID string) (*pubsub.Topic, error) {
	client, err := pubsub.NewClient(ctx, "test_project")
	if err != nil {
		return nil, err
	}

	log.Println("Pub/Sub client created")
	log.Printf("%+v", client)

	defer client.Close()

	topic := client.Topic(topicID)

	// Create the topic if it doesn't exist.
	// exists, err := topic.Exists(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	// if !exists {
	log.Printf("Topic %v doesn't exist - creating it", topicID)

	if _, err = client.CreateTopic(ctx, topicID); err != nil {
		return nil, err
	}
	// }

	log.Printf("%+v", topic)
	return topic, nil
}

func SendMessageToPubSub(ctx context.Context, uri string, topic *pubsub.Topic) error {
	msg := &pubsub.Message{
		Data: []byte("This is a test message!"),
	}

	log.Printf("%+v", topic)
	log.Println("About to publish message!")

	if _, err := topic.Publish(ctx, msg).Get(ctx); err != nil {
		return err
	}

	log.Println("Message published.")
	return nil
}

func CheckMessageReceived(ctx context.Context, uri string, topic *pubsub.Topic, subscriptionID string) error {
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
