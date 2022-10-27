package sakerhet_test

// Basic imports
import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	abstractedcontainers "github.com/averageflow/sakerhet/pkg/abstracted_containers"
	"github.com/averageflow/sakerhet/pkg/sakerhet"
	"github.com/google/uuid"
)

func TestHighLevelIntegrationTestGCPPubSub(t *testing.T) {
	// t.Parallel() not possible yet due to os.Setenv("PUBSUB_EMULATOR_HOST", pubSubC.URI)

	// given
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*15))
	defer cancel()

	i := sakerhet.IntegrationTest{
		GCPPubSubData: &sakerhet.IntegrationTestGCPPubSubData{
			ProjectID:      "test-project",
			TopicID:        "test-topic-" + uuid.New().String(),
			SubscriptionID: "test-sub-" + uuid.New().String(),
		},
		TestContext: ctx,
		T:           t,
	}

	pubSubC, err := i.GCPPubSubContainerStart()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = pubSubC.Terminate(i.TestContext)
	}()

	wantedData := []byte(`{"myKey": "myValue"}`)

	// when
	if err := i.GCPPubSubPublishData(wantedData); err != nil {
		t.Fatal(err)
	}

	// then
	expectedData := [][]byte{[]byte(wantedData)}
	if err := i.GCPPubSubContainsWantedMessages(
		1000*time.Millisecond,
		expectedData,
	); err != nil {
		t.Fatal(err)
	}
}

func TestLowLevelIntegrationTestGCPPubSub(t *testing.T) {
	// t.Parallel() not possible yet due to os.Setenv("PUBSUB_EMULATOR_HOST", pubSubC.URI)

	projectID := "test-project"
	topicID := "test-topic-" + uuid.New().String()
	subscriptionID := "test-sub-" + uuid.New().String()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*15))
	defer cancel()

	topicSubscriptionMap := map[string][]string{
		topicID: {subscriptionID},
	}

	pubSubC, err := abstractedcontainers.SetupGCPPubsub(ctx, projectID, topicSubscriptionMap)
	if err != nil {
		t.Error(err)
	}

	// clean up the container after the test is complete
	defer func() {
		_ = pubSubC.Terminate(ctx)
	}()

	fmt.Printf("New container started, accessible at: %s\n", pubSubC.URI)

	// required so that all Pub/Sub calls go to docker container, and not GCP
	os.Setenv("PUBSUB_EMULATOR_HOST", pubSubC.URI)

	client, err := pubsub.NewClient(context.TODO(), projectID)
	if err != nil {
		t.Fatal(err)
	}

	defer client.Close()

	topic, err := abstractedcontainers.GetOrCreateGCPTopic(ctx, client, topicID)
	if err != nil {
		t.Fatal(err)
	}

	wantedData := []byte(`{"myKey": "myValue"}`)

	if err := abstractedcontainers.PublishToGCPTopic(ctx, client, topic, wantedData); err != nil {
		t.Fatal(err)
	}

	expectedData := [][]byte{[]byte(wantedData)}
	if err := abstractedcontainers.CheckGCPMessageInSub(ctx, client, subscriptionID, expectedData, 1000*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}
