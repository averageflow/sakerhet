package abstractedcontainers_test

// Basic imports
import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	abstractedcontainers "github.com/averageflow/sakerhet/internal/abstracted_containers"
	"github.com/google/uuid"
)

func TestExample(t *testing.T) {
	topicID := "topic-" + uuid.New().String()
	subscriptionID := "sub-" + uuid.New().String()

	log.Printf("topic: %s , subscription: %s", topicID, subscriptionID)

	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)

	pubSubC, err := abstractedcontainers.SetupPubsub(ctx)
	if err != nil {
		t.Error(err)
	}

	os.Setenv("PUBSUB_EMULATOR_HOST", pubSubC.URI)

	// Clean up the container after the test is complete
	defer pubSubC.Terminate(ctx)

	topic, err := abstractedcontainers.GetOrCreateTopic(ctx, pubSubC.URI, topicID)
	if err != nil {
		t.Error(err)
		return
	}

	if err := abstractedcontainers.SendMessageToPubSub(ctx, pubSubC.URI, topic); err != nil {
		t.Error(err)
		return
	}

	if err := abstractedcontainers.CheckMessageReceived(ctx, pubSubC.URI, topic, subscriptionID); err != nil {
		t.Error(err)
		return
	}
}
