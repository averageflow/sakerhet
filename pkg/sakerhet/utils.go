package sakerhet

import (
	"fmt"
	"os"
	"testing"
)

const (
	SakerhetRunIntegrationTestsEnvVar = "SAKERHET_RUN_INTEGRATION_TESTS"
)

func UnorderedEqualByteArrays(first, second [][]byte) bool {
	var firstA, secondA []any

	for _, v := range first {
		firstA = append(firstA, v)
	}

	for _, v := range second {
		secondA = append(secondA, v)
	}

	return UnorderedEqual(firstA, secondA)
}

func UnorderedEqual(first, second []any) bool {
	if len(first) != len(second) {
		return false
	}

	exists := make(map[string]bool)

	for _, value := range first {
		exists[fmt.Sprintf("%+v", value)] = true
	}

	for _, value := range second {
		if !exists[fmt.Sprintf("%+v", value)] {
			return false
		}
	}

	return true
}

func SkipUnitTestsWhenIntegrationTesting(t *testing.T) {
	if os.Getenv(SakerhetRunIntegrationTestsEnvVar) != "" {
		t.Skip("Skipping unit tests! Unset variable SAKERHET_RUN_INTEGRATION_TESTS to run them!")
	}
}

func SkipIntegrationTestsWhenUnitTesting(t *testing.T) {
	if os.Getenv(SakerhetRunIntegrationTestsEnvVar) == "" {
		t.Skip("Skipping integration tests! Set variable SAKERHET_RUN_INTEGRATION_TESTS to run them!")
	}
}
