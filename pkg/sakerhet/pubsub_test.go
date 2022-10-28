package sakerhet_test

// Basic imports
import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	abstractedcontainers "github.com/averageflow/sakerhet/pkg/abstracted_containers"
	"github.com/averageflow/sakerhet/pkg/sakerhet"
	"github.com/google/uuid"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// High level test on code that pushes to Pub/Sub
func TestHighLevelIntegrationTestGCPPubSub(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*30))
	defer cancel()

	i := sakerhet.NewIntegrationTest(sakerhet.IntegrationTestParams{
		TestContext: ctx,
		GCPPubSub:   &sakerhet.GCPPubSubIntegrationTestParams{},
	})

	pubSubC, err := i.GCPPubSubIntegrationTester.ContainerStart()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = pubSubC.Terminate(ctx)
	}()

	wantedData := []byte(`{"myKey": "myValue"}`)

	// when
	if err := i.GCPPubSubIntegrationTester.PublishData(wantedData); err != nil {
		t.Fatal(err)
	}

	// then
	expectedData := [][]byte{[]byte(wantedData)}
	if err := i.GCPPubSubIntegrationTester.ContainsWantedMessages(
		3*time.Second,
		expectedData,
	); err != nil {
		t.Fatal(err)
	}
}

// simple service that does a computation and publishes to Pub/Sub
type myPowerOfNService struct {
	toPowerOfN    func(float64, float64) float64
	publishResult func(context.Context, string, string, string, float64) error
}

func newMyPowerOfNService() myPowerOfNService {
	return myPowerOfNService{
		toPowerOfN: func(x float64, n float64) float64 { return math.Pow(x, n) },
		publishResult: func(ctx context.Context, pubSubURI, projectID, topicID string, x float64) error {
			conn, err := grpc.Dial(pubSubURI, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return fmt.Errorf("grpc.Dial: %v", err)
			}

			o := []option.ClientOption{
				option.WithGRPCConn(conn),
				option.WithTelemetryDisabled(),
			}

			client, err := pubsub.NewClientWithConfig(ctx, projectID, nil, o...)
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
	t.Parallel()

	// given
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*30))
	defer cancel()

	i := sakerhet.NewIntegrationTest(sakerhet.IntegrationTestParams{
		TestContext: ctx,
		GCPPubSub:   &sakerhet.GCPPubSubIntegrationTestParams{},
	})

	pubSubC, err := i.GCPPubSubIntegrationTester.ContainerStart()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = pubSubC.Terminate(ctx)
	}()

	// when
	ponService := newMyPowerOfNService()

	if err := ponService.publishResult(
		ctx,
		i.GCPPubSubIntegrationTester.PubSubURI,
		i.GCPPubSubIntegrationTester.ProjectID,
		i.GCPPubSubIntegrationTester.TopicID,
		ponService.toPowerOfN(3, 3),
	); err != nil {
		t.Fatal(err)
	}

	// then
	expectedData := [][]byte{[]byte(`{"computationResult": 27.00}`)}
	if err := i.GCPPubSubIntegrationTester.ContainsWantedMessages(
		3*time.Second,
		expectedData,
	); err != nil {
		t.Fatal(err)
	}
}

// Low level test with full control on testing code that pushes to Pub/Sub
func TestLowLevelIntegrationTestGCPPubSub(t *testing.T) {
	t.Parallel()

	// given
	projectID := "test-project"
	topicID := "test-topic-" + uuid.New().String()
	subscriptionID := "test-sub-" + uuid.New().String()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*30))
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
	// os.Setenv("PUBSUB_EMULATOR_HOST", pubSubC.URI)
	conn, err := grpc.Dial(pubSubC.URI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(fmt.Errorf("grpc.Dial: %v", err))
	}

	o := []option.ClientOption{
		option.WithGRPCConn(conn),
		option.WithTelemetryDisabled(),
	}

	client, err := pubsub.NewClientWithConfig(ctx, projectID, nil, o...)
	if err != nil {
		t.Fatal(err)
	}

	defer client.Close()

	topic, err := abstractedcontainers.GetOrCreateGCPTopic(ctx, client, topicID)
	if err != nil {
		t.Fatal(err)
	}

	// when
	wantedData := []byte(`{"myKey": "myValue"}`)

	if err := abstractedcontainers.PublishToGCPTopic(ctx, client, topic, wantedData); err != nil {
		t.Fatal(err)
	}

	// theNn
	expectedData := [][]byte{[]byte(wantedData)}
	if err := abstractedcontainers.CheckGCPMessageInSub(ctx, client, subscriptionID, expectedData, 3*time.Second); err != nil {
		t.Fatal(err)
	}
}
