package abstractedcontainers_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAbstractedContainers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AbstractedContainers Suite")
}
