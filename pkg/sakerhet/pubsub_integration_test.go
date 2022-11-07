package sakerhet_test

// Basic imports
import (
	"context"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	abstractedcontainers "github.com/averageflow/sakerhet/pkg/abstracted_containers"
	dummyservices "github.com/averageflow/sakerhet/pkg/dummy_services"
	"github.com/averageflow/sakerhet/pkg/sakerhet"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Test suite demonstrating the use of Säkerhet + testcontainers for GCP Pub/Sub
type GCPPubSubTestSuite struct {
	suite.Suite
	TestContext        context.Context
	TestContextCancel  context.CancelFunc
	GCPPubSubContainer *abstractedcontainers.GCPPubSubContainer
	IntegrationTester  sakerhet.Sakerhet
}

// Before suite starts
func (suite *GCPPubSubTestSuite) SetupSuite() {
	ctx := context.Background()

	suite.IntegrationTester = sakerhet.NewSakerhetIntegrationTest(sakerhet.SakerhetBuilder{
		GCPPubSub: &sakerhet.GCPPubSubIntegrationTestParams{},
	})

	// Spin up one GCP Pub/Sub container for all the tests in the suite
	pubSubContainer, err := suite.IntegrationTester.GCPPubSubIntegrationTester.ContainerStart(ctx)
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.GCPPubSubContainer = pubSubContainer
}

// Before each test
func (suite *GCPPubSubTestSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), sakerhet.GetIntegrationTestTimeout())
	suite.TestContext = ctx
	suite.TestContextCancel = cancel
}

// After suite ends
func (suite *GCPPubSubTestSuite) TearDownSuite() {
	// Spin down the container
	_ = suite.GCPPubSubContainer.Terminate(context.Background())
}

// Start the test suite if we are running integration tests
func TestGCPPubSubTestSuite(t *testing.T) {
	sakerhet.SkipIntegrationTestsWhenUnitTesting(t)
	suite.Run(t, new(GCPPubSubTestSuite))
}

// High level test on code that pushes to Pub/Sub
func (suite *GCPPubSubTestSuite) TestHighLevelIntegrationTestGCPPubSub() {
	// given
	payload := []byte(`{"myKey": "myValue"}`)

	// when
	if err := suite.IntegrationTester.GCPPubSubIntegrationTester.PublishData(suite.TestContext, payload); err != nil {
		suite.T().Fatal(err)
	}

	// then
	expectedData := [][]byte{[]byte(payload)}
	if err := suite.IntegrationTester.GCPPubSubIntegrationTester.ContainsWantedMessages(
		suite.TestContext,
		expectedData,
	); err != nil {
		suite.T().Fatal(err)
	}
}

// High level test of a service that publishes to Pub/Sub
func (suite *GCPPubSubTestSuite) TestHighLevelIntegrationTestOfServiceThatUsesGCPPubSub() {
	// given
	powerOfNService := dummyservices.NewMyPowerOfNService()

	// when
	x := powerOfNService.ToPowerOfN(3, 3)
	y := powerOfNService.ToPowerOfN(4, 2)

	if err := suite.IntegrationTester.GCPPubSubIntegrationTester.PublishData(
		suite.TestContext,
		[]byte(fmt.Sprintf(`{"computationResult": %.2f}`, x)),
	); err != nil {
		suite.T().Fatal(err)
	}

	if err := suite.IntegrationTester.GCPPubSubIntegrationTester.PublishData(
		suite.TestContext,
		[]byte(fmt.Sprintf(`{"computationResult": %.2f}`, y)),
	); err != nil {
		suite.T().Fatal(err)
	}

	// then
	expectedData := [][]byte{[]byte(`{"computationResult": 27.00}`), []byte(`{"computationResult": 16.00}`)}
	if err := suite.IntegrationTester.GCPPubSubIntegrationTester.ContainsWantedMessages(
		suite.TestContext,
		expectedData,
	); err != nil {
		suite.T().Fatal(err)
	}
}

// Low level test with full control on testing code that pushes to Pub/Sub
func TestLowLevelIntegrationTestGCPPubSub(t *testing.T) {
	sakerhet.SkipIntegrationTestsWhenUnitTesting(t)

	// given
	projectID := "test-project"
	topicID := "test-topic-" + uuid.New().String()
	subscriptionID := "test-sub-" + uuid.New().String()

	topicSubscriptionMap := map[string][]string{
		topicID: {subscriptionID},
	}

	pubSubContainer, err := abstractedcontainers.SetupGCPPubsub(context.Background(), projectID, topicSubscriptionMap)
	if err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), sakerhet.GetIntegrationTestTimeout())
	defer cancel()

	// clean up the container after the test is complete
	defer func() {
		_ = pubSubContainer.Terminate(context.Background())
	}()

	conn, err := grpc.Dial(pubSubContainer.URI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(fmt.Errorf("grpc.Dial: %v", err))
	}

	gcpPubSubClientOptions := []option.ClientOption{
		option.WithGRPCConn(conn),
		option.WithTelemetryDisabled(),
	}

	client, err := pubsub.NewClientWithConfig(ctx, projectID, nil, gcpPubSubClientOptions...)
	if err != nil {
		t.Fatal(err)
	}

	defer client.Close()

	topic, err := sakerhet.GetOrCreateGCPTopic(ctx, client, topicID)
	if err != nil {
		t.Fatal(err)
	}

	// when
	payload := []byte(`{"myKey": "myValue"}`)
	if err := sakerhet.PublishToGCPTopic(ctx, client, topic, payload); err != nil {
		t.Fatal(err)
	}

	// then
	expectedData := [][]byte{[]byte(payload)}
	if err := sakerhet.ExpectGCPMessagesInSub(ctx, client, subscriptionID, expectedData, 1500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}
