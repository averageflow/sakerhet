package abstractedcontainers_test

// Basic imports
import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	abstractedcontainers "github.com/averageflow/sakerhet/internal/abstracted_containers"
)

// func TestWithRedis(t *testing.T) {
// 	ctx := context.Background()
// 	req := testcontainers.ContainerRequest{
// 		Image:        "redis:latest",
// 		ExposedPorts: []string{"6379/tcp"},
// 		WaitingFor:   wait.ForLog("Ready to accept connections"),
// 	}
// 	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
// 		ContainerRequest: req,
// 		Started:          true,
// 	})
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	defer redisC.Terminate(ctx)
// }

func TestExample(t *testing.T) {
	// topicID := "topic-" + uuid.New().String()
	// subscriptionID := "sub-" + uuid.New().String()
	topicID := "myTopic"
	subscriptionID := "mySub"

	fmt.Printf("topic: %s , subscription: %s", topicID, subscriptionID)

	ctx, _ := context.WithTimeout(context.Background(), 2000*time.Second)

	pubSubC, err := abstractedcontainers.SetupPubsub(ctx)
	if err != nil {
		t.Error(err)
	}

	os.Setenv("PUBSUB_EMULATOR_HOST", pubSubC.URI)
	// os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")

	// Clean up the container after the test is complete
	defer pubSubC.Terminate(ctx)

	topic, err := abstractedcontainers.GetOrCreateTopic(ctx, pubSubC.URI, topicID)
	// topic, err := abstractedcontainers.GetOrCreateTopic(ctx, "localhost:8085", topicID)
	if err != nil {
		t.Error(err)
		return
	}

	if err := abstractedcontainers.SendMessageToPubSub(ctx, pubSubC.URI, topic); err != nil {
		// if err := abstractedcontainers.SendMessageToPubSub(ctx, "localhost:8085", topic); err != nil {
		t.Error(err)
		return
	}

	if err := abstractedcontainers.CheckMessageReceived(ctx, pubSubC.URI, topic, subscriptionID); err != nil {
		// if err := abstractedcontainers.CheckMessageReceived(ctx, "localhost:8085", topic, subscriptionID); err != nil {
		t.Error(err)
		return
	}
}
