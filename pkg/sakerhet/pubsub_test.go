package sakerhet_test

// Basic imports
import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	abstractedcontainers "github.com/averageflow/sakerhet/pkg/abstracted_containers"
	"github.com/averageflow/sakerhet/pkg/sakerhet"
	"github.com/google/uuid"
)

// High level test on code that pushes to Pub/Sub
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

// Low level test with full control on testing code that pushes to Pub/Sub
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

// simple service that does a computation and publishes to Pub/Sub
type myPowerOfNService struct {
	toPowerOfN    func(float64, float64) float64
	publishResult func(context.Context, string, string, float64) error
}

func NewMyPowerOfNService() myPowerOfNService {
	return myPowerOfNService{
		toPowerOfN: func(x float64, n float64) float64 { return math.Pow(x, n) },
		publishResult: func(ctx context.Context, projectID, topicID string, x float64) error {
			client, err := pubsub.NewClient(ctx, projectID)
			if err != nil {
				return err
			}

			defer client.Close()

			topic, err := abstractedcontainers.GetOrCreateGCPTopic(ctx, client, topicID)
			if err != nil {
				return err
			}

			payloadToPublish := []byte(fmt.Sprintf(`{"computationResult": %.2f}`, x))

			if err := abstractedcontainers.PublishToGCPTopic(ctx, client, topic, payloadToPublish); err != nil {
				return err
			}

			return nil
		},
	}
}

// High level test of a service that publishes to Pub/Sub
func TestHighLevelIntegrationTestOfServiceThatUsesGCPPubSub(t *testing.T) {
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

	// when
	ponService := NewMyPowerOfNService()

	if err := ponService.publishResult(
		i.TestContext,
		i.GCPPubSubData.ProjectID,
		i.GCPPubSubData.TopicID,
		ponService.toPowerOfN(3, 3),
	); err != nil {
		t.Fatal(err)
	}

	// then
	expectedData := [][]byte{[]byte(`{"computationResult": 27.00}`)}
	if err := i.GCPPubSubContainsWantedMessages(
		1000*time.Millisecond,
		expectedData,
	); err != nil {
		t.Fatal(err)
	}
}
