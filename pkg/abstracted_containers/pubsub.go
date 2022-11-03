package abstractedcontainers

import (
	"context"
	"fmt"

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
		Image: "thekevjames/gcloud-pubsub-emulator:406.0.0",
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
		AutoRemove: true,
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
