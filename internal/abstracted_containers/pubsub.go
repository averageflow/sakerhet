package abstractedcontainers

import (
	"context"
	"fmt"
	"sync/atomic"
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
		Image:        "thekevjames/gcloud-pubsub-emulator:latest",
		ExposedPorts: []string{"8682/tcp", "8681/tcp"},
		Env: map[string]string{
			// "PUBSUB_PROJECT1": "PROJECTID,TOPIC1,TOPIC2:SUBSCRIPTION1:SUBSCRIPTION2,TOPIC3:SUBSCRIPTION3"
			// "PUBSUB_PROJECT1": "myProject,myTopic:mySub",
			"PUBSUB_PROJECT1": "myProject,myTopic",
		},
		//Hostname:     "0.0.0.0",
		//WaitingFor:   wait.ForLog("[pubsub] INFO: Server started, listening on 8538"),
		//NetworkMode:  "bridge",
		WaitingFor: wait.NewHostPortStrategy("8682"),
		//Cmd:        []string{"gcloud beta emulators pubsub start --project=test_project --host-port=0.0.0.0:8085"},
		// Cmd: []string{"gcloud beta emulators pubsub start --project=test_project --host-port=0.0.0.0:8085"},
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

	mappedPort, err := container.MappedPort(ctx, "8681")
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s:%s", ip, mappedPort.Port())
	// uri := fmt.Sprintf("172.17.0.1:%s", mappedPort.Port())
	fmt.Printf("New container started, accessible at: %s", uri)

	return &PubSubContainer{Container: container, URI: uri}, nil
}

func GetOrCreateTopic(ctx context.Context, uri, topicID string) (*pubsub.Topic, error) {
	client, err := pubsub.NewClient(context.TODO(), "myProject")
	if err != nil {
		return nil, err
	}

	fmt.Println("Pub/Sub client created")
	fmt.Printf("%+v", client)

	defer client.Close()

	topic := client.Topic(topicID)

	// Create the topic if it doesn't exist.
	exists, err := topic.Exists(context.TODO())
	if err != nil {
		return nil, err
	}

	if !exists {
		fmt.Printf("Topic %v doesn't exist - creating it", topicID)

		if _, err = client.CreateTopic(context.TODO(), topicID); err != nil {
			return nil, err
		}
	} else {
		fmt.Println("Not required to create topic, it exists")
	}

	return topic, nil
}

func SendMessageToPubSub(ctx context.Context, uri string, topic *pubsub.Topic) error {
	msg := &pubsub.Message{
		Data: []byte("This is a test message!"),
	}

	fmt.Println("About to publish message!")

	// if _, err := topic.Publish(context.TODO(), msg).Get(context.TODO()); err != nil {
	// 	return err
	// }

	res := topic.Publish(context.TODO(), msg)
	// if _, err := .Get(context.TODO()); err != nil {
	// 	return err
	// }

	fmt.Printf("%+v", res)

	fmt.Println("Message published.")
	return nil
}

func CheckMessageReceived(ctx context.Context, uri string, topic *pubsub.Topic, subscriptionID string) error {
	client, err := pubsub.NewClient(context.TODO(), "myProject")
	if err != nil {
		return err
	}

	defer client.Close()

	sub, err := client.CreateSubscription(ctx, subscriptionID, pubsub.SubscriptionConfig{
		Topic: topic,
	})

	// sub := client.Subscription("mySub")

	// Receive messages for 10 seconds, which simplifies testing.
	// Comment this out in production, since `Receive` should
	// be used as a long running operation.
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var received int32

	err = sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		fmt.Printf("Got message: %q\n", string(msg.Data))
		atomic.AddInt32(&received, 1)
		msg.Ack()
	})
	if err != nil {
		return fmt.Errorf("sub.Receive: %v", err)
	}

	fmt.Printf("Received %d messages\n", received)

	return nil

	// cctx, cancel := context.WithCancel(context.TODO())
	// okCh := make(chan string)

	// go func() {
	// 	// Use a callback to receive messages via subscription.
	// 	// Receive will block until the context is cancelled, or we get a non-recoverable error
	// 	err = sub.Receive(cctx, func(_ context.Context, m *pubsub.Message) {
	// 		m.Ack()
	// 		okCh <- string(m.Data)
	// 		cancel()
	// 	})
	// }()

	// select {
	// case msg := <-okCh:
	// 	if msg != "This is a test message!" {
	// 		return fmt.Errorf("")
	// 	}
	// case <-time.After(300000 * time.Millisecond):
	// 	cancel()
	// 	return fmt.Errorf("did not receive message within deadline")
	// }

	// return nil
}
