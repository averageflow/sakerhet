//go:build integration

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
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GCPPubSubTestSuite struct {
	suite.Suite
	TestContext        context.Context
	TestContextCancel  context.CancelFunc
	GCPPubSubContainer *abstractedcontainers.GCPPubSubContainer
	IntegrationTester  sakerhet.IntegrationTester
}

// before each test
func (suite *GCPPubSubTestSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*30))
	suite.TestContext = ctx
	suite.TestContextCancel = cancel

	suite.IntegrationTester = sakerhet.NewIntegrationTest(sakerhet.IntegrationTesterParams{
		TestContext: ctx,
		GCPPubSub:   &sakerhet.GCPPubSubIntegrationTestParams{},
	})

	pubSubC, err := suite.IntegrationTester.GCPPubSubIntegrationTester.ContainerStart()
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.GCPPubSubContainer = pubSubC
}

func (suite *GCPPubSubTestSuite) TearDownSuite() {
	_ = suite.GCPPubSubContainer.Terminate(suite.TestContext)
}

func TestGCPPubSubTestSuite(t *testing.T) {
	suite.Run(t, new(GCPPubSubTestSuite))
}

// High level test on code that pushes to Pub/Sub
func (suite *GCPPubSubTestSuite) TestHighLevelIntegrationTestGCPPubSub() {
	wantedData := []byte(`{"myKey": "myValue"}`)

	// when
	if err := suite.IntegrationTester.GCPPubSubIntegrationTester.PublishData(wantedData); err != nil {
		suite.T().Fatal(err)
	}

	// then
	expectedData := [][]byte{[]byte(wantedData)}
	if err := suite.IntegrationTester.GCPPubSubIntegrationTester.ContainsWantedMessages(
		1*time.Second,
		expectedData,
	); err != nil {
		suite.T().Fatal(err)
	}
}

// High level test of a service that publishes to Pub/Sub
func (suite *GCPPubSubTestSuite) TestHighLevelIntegrationTestOfServiceThatUsesGCPPubSub() {
	// simple service that does a computation and publishes to Pub/Sub
	type myPowerOfNService struct {
		toPowerOfN func(float64, float64) float64
	}

	newMyPowerOfNService := func() myPowerOfNService {
		return myPowerOfNService{
			toPowerOfN: func(x float64, n float64) float64 { return math.Pow(x, n) },
		}
	}

	// when
	ponService := newMyPowerOfNService()
	x := ponService.toPowerOfN(3, 3)
	y := ponService.toPowerOfN(4, 2)

	if err := suite.IntegrationTester.GCPPubSubIntegrationTester.PublishData(
		[]byte(fmt.Sprintf(`{"computationResult": %.2f}`, x)),
	); err != nil {
		suite.T().Fatal(err)
	}

	if err := suite.IntegrationTester.GCPPubSubIntegrationTester.PublishData(
		[]byte(fmt.Sprintf(`{"computationResult": %.2f}`, y)),
	); err != nil {
		suite.T().Fatal(err)
	}

	// then
	expectedData := [][]byte{[]byte(`{"computationResult": 27.00}`), []byte(`{"computationResult": 16.00}`)}
	if err := suite.IntegrationTester.GCPPubSubIntegrationTester.ContainsWantedMessages(
		1*time.Second,
		expectedData,
	); err != nil {
		suite.T().Fatal(err)
	}
}

// Low level test with full control on testing code that pushes to Pub/Sub
func TestLowLevelIntegrationTestGCPPubSub(t *testing.T) {
	// t.Parallel()

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

	// then
	expectedData := [][]byte{[]byte(wantedData)}
	if err := abstractedcontainers.CheckGCPMessageInSub(ctx, client, subscriptionID, expectedData, 1*time.Second); err != nil {
		t.Fatal(err)
	}
}
