package abstractedcontainers_test

// Basic imports
import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

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

	fmt.Printf("New container started, accessible at: %s\n", pubSubC.URI)

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

	wantedData := []byte(`{"myKey": "myValue"}`)

	if err := abstractedcontainers.PublishToGCPTopic(ctx, client, topic, wantedData); err != nil {
		t.Error(err)
		return
	}

	expectedData := [][]byte{[]byte(wantedData)}
	if err := abstractedcontainers.CheckGCPMessageInSub(ctx, client, subscriptionID, expectedData, 100*time.Millisecond); err != nil {
		t.Error(err)
		return
	}
}
