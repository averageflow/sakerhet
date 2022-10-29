package abstractedcontainers

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type GCPPubSubContainer struct {
	testcontainers.Container
	URI               string
	LivenessProbePort nat.Port
	PubSubPort        nat.Port
}

func serializeTopicSubscriptionMapForDockerEnv(topicSubscriptionMap map[string][]string) string {
	var serialized string

	for i, v := range topicSubscriptionMap {
		serialized += i

		for _, vv := range v {
			serialized += fmt.Sprintf(":%s", vv)
		}
	}

	return serialized
}

func SetupGCPPubsub(ctx context.Context, projectID string, topicSubscriptionMap map[string][]string) (*GCPPubSubContainer, error) {
	livenessProbePort, err := nat.NewPort("tcp", "8682")
	if err != nil {
		return nil, err
	}

	pubSubPort, err := nat.NewPort("tcp", "8681")
	if err != nil {
		return nil, err
	}

	req := testcontainers.ContainerRequest{
		Image: "thekevjames/gcloud-pubsub-emulator:latest",
		ExposedPorts: []string{
			fmt.Sprintf("%s/%s", livenessProbePort.Port(), livenessProbePort.Proto()),
			fmt.Sprintf("%s/%s", pubSubPort.Port(), pubSubPort.Proto()),
		},
		Env: map[string]string{
			// specify the topics and subscriptions to be created, in the docker container's environment variable
			// "PUBSUB_PROJECT1": "PROJECTID,TOPIC1,TOPIC2:SUBSCRIPTION1:SUBSCRIPTION2,TOPIC3:SUBSCRIPTION3"
			"PUBSUB_PROJECT1": fmt.Sprintf("%s,%s", projectID, serializeTopicSubscriptionMapForDockerEnv(topicSubscriptionMap)),
		},
		// await until communication is possible on liveness probe port, then proceed
		WaitingFor: wait.ForListeningPort(livenessProbePort),
		Name:       "gcp-pubsub",
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            true,
	})
	if err != nil {
		return nil, err
	}

	ip, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, pubSubPort)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s:%s", ip, mappedPort.Port())

	return &GCPPubSubContainer{
		Container:         container,
		URI:               uri,
		LivenessProbePort: livenessProbePort,
		PubSubPort:        pubSubPort,
	}, nil
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

func PublishToGCPTopic(ctx context.Context, client *pubsub.Client, topic *pubsub.Topic, wantedData []byte) error {
	var wg sync.WaitGroup
	var totalErrors uint64

	result := topic.Publish(ctx, &pubsub.Message{
		Data: wantedData,
	})

	wg.Add(1)
	go func(res *pubsub.PublishResult) {
		defer wg.Done()
		// The Get method blocks until a server-generated ID or
		// an error is returned for the published message.
		id, err := res.Get(ctx)
		if err != nil {
			// Error handling code can be added here.
			fmt.Printf("Failed to publish: %v", err)
			atomic.AddUint64(&totalErrors, 1)
			return
		}
		fmt.Printf("Published message %d; msg ID: %v\n", 1, id)
	}(result)

	wg.Wait()

	if totalErrors > 0 {
		return fmt.Errorf("%d messages did not publish successfully", totalErrors)
	}
	return nil
}

func toReadableSliceOfByteSlices(raw [][]byte) []string {
	result := make([]string, len(raw))

	for i := range raw {
		result[i] = string(raw[i])
	}

	return result
}

// Receive messages for 3 seconds, which simplifies testing.
func CheckGCPMessageInSub(ctx context.Context, client *pubsub.Client, subscriptionID string, wantedData [][]byte, timeToWait time.Duration) error {
	sub := client.Subscription(subscriptionID)

	ctx, cancel := context.WithTimeout(ctx, timeToWait)
	defer cancel()

	var receivedData [][]byte

	err := sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		fmt.Printf("Got message: %q\n", string(msg.Data))
		receivedData = append(receivedData, msg.Data)
		msg.Ack()
		// cancel()
	})
	if err != nil {
		return fmt.Errorf("sub.Receive: %v", err)
	}

	fmt.Printf("Received %d messages\n", len(receivedData))

	if !reflect.DeepEqual(wantedData, receivedData) {
		return fmt.Errorf(
			"received data is different than expected:\n received %v\n expected %v\n",
			toReadableSliceOfByteSlices(receivedData),
			toReadableSliceOfByteSlices(wantedData),
		)
	}

	return nil
}
