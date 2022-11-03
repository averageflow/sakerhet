package sakerhet_test

import (
	"testing"

	"github.com/averageflow/sakerhet/pkg/sakerhet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UnorderedEqualTestSuite struct {
	suite.Suite
}

func TestUnorderedEqualTestSuite(t *testing.T) {
	sakerhet.SkipUnitTestsWhenIntegrationTesting(t)
	t.Parallel()
	suite.Run(t, new(UnorderedEqualTestSuite))
}

func (suite *UnorderedEqualTestSuite) TestUnorderedEqual() {
	assert.True(suite.T(), sakerhet.UnorderedEqual(
		[]string{"one", "two"},
		[]string{"two", "one"},
	))

	assert.True(suite.T(), sakerhet.UnorderedEqual(
		[]bool{true, true, false},
		[]bool{false, true, true},
	))

	assert.False(suite.T(), sakerhet.UnorderedEqual(
		[]int{1, 3},
		[]int{2, 3},
	))

	assert.False(suite.T(), sakerhet.UnorderedEqual(
		[]string{"one"},
		[]string{"two"},
	))

	assert.True(suite.T(), sakerhet.UnorderedEqual(
		[][]byte{[]byte(`{"foo": "bar"}`), []byte(`!@#$%^&*()`)},
		[][]byte{[]byte(`!@#$%^&*()`), []byte(`{"foo": "bar"}`)},
	))

	assert.False(suite.T(), sakerhet.UnorderedEqual(
		[][]byte{[]byte(`someBytes`)},
		[][]byte{[]byte(`someBytes`), []byte(`moreBytes`)},
	))
}
