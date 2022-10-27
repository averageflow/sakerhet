package abstractedcontainers_test

// Basic imports
import (
	"context"
	"fmt"
	"os"
	"testing"

	"cloud.google.com/go/pubsub"
	abstractedcontainers "github.com/averageflow/sakerhet/internal/abstracted_containers"
	"github.com/google/uuid"
)

func TestPubSub(t *testing.T) {
	projectID := "test-project"
	topicID := "test-topic-" + uuid.New().String()
	subscriptionID := "test-sub-" + uuid.New().String()

	ctx := context.TODO()

	topicSubscriptionMap := map[string][]string{
		topicID: {subscriptionID},
	}

	pubSubC, err := abstractedcontainers.SetupGCPPubsub(ctx, projectID, topicSubscriptionMap)
	if err != nil {
		t.Error(err)
	}

	// clean up the container after the test is complete
	defer pubSubC.Terminate(ctx)

	fmt.Printf("New container started, accessible at: %s", pubSubC.URI)

	// required so that all Pub/Sub calls go to docker container, and not GCP
	os.Setenv("PUBSUB_EMULATOR_HOST", pubSubC.URI)

	client, err := pubsub.NewClient(context.TODO(), projectID)
	if err != nil {
		t.Error(err)
		return
	}

	defer client.Close()

	topic, err := abstractedcontainers.GetOrCreateGCPTopic(ctx, client, topicID)
	if err != nil {
		t.Error(err)
		return
	}

	if err := abstractedcontainers.PublishToGCPTopic(ctx, client, topic); err != nil {
		t.Error(err)
		return
	}

	if err := abstractedcontainers.CheckGCPMessageInSub(ctx, client, subscriptionID); err != nil {
		t.Error(err)
		return
	}
}
