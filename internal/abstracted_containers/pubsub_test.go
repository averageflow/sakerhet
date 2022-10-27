package abstractedcontainers_test

import (
	"context"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"

	//. "github.com/onsi/gomega"

	abstractedcontainers "github.com/averageflow/sakerhet/internal/abstracted_containers"
)

var _ = Describe("Pubsub", func() {
	topicID := "topic-" + uuid.New().String()
	subscriptionID := "sub-" + uuid.New().String()

	ctx := context.Background()

	pubSubC, err := abstractedcontainers.SetupPubsub(ctx)
	if err != nil {
		panic(err.Error())
	}

	// Clean up the container after the test is complete
	defer pubSubC.Terminate(ctx)

	It("can publish to the Pub/Sub container without errors", func() {
		topic, err := abstractedcontainers.GetOrCreateTopic(pubSubC.URI, topicID)
		if err != nil {
			Fail(err.Error())
		}

		if err := abstractedcontainers.SendMessageToPubSub(pubSubC.URI, topic); err != nil {
			Fail(err.Error())
		}

		if err := abstractedcontainers.CheckMessageReceived(pubSubC.URI, topic, subscriptionID); err != nil {
			Fail(err.Error())
		}
	})

	// BeforeEach(func() {
	//
	// })
})
