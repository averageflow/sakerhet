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
	type unorderedEqualTestCase struct {
		left            []any
		right           []any
		expectToBeEqual bool
	}

	sut := []unorderedEqualTestCase{
		{
			left:            []any{"one", "two"},
			right:           []any{"two", "one"},
			expectToBeEqual: true,
		},
		{
			left:            []any{true, true, false},
			right:           []any{false, true, true},
			expectToBeEqual: true,
		},
		{
			left:            []any{1, 2, 3},
			right:           []any{3, 2, 1},
			expectToBeEqual: true,
		},
		{
			left:            []any{1, 3},
			right:           []any{2, 3},
			expectToBeEqual: false,
		},
		{
			left:            []any{"one"},
			right:           []any{"two"},
			expectToBeEqual: false,
		},
	}

	for _, v := range sut {
		areEqual := sakerhet.UnorderedEqual(v.left, v.right)
		assert.Equal(suite.T(), v.expectToBeEqual, areEqual)
	}
}

func (suite *UnorderedEqualTestSuite) TestUnorderedEqualByteArrays() {
	type unorderedEqualTestCase struct {
		left            [][]byte
		right           [][]byte
		expectToBeEqual bool
	}

	sut := []unorderedEqualTestCase{
		{
			left:            [][]byte{[]byte(`{"foo": "bar"}`), []byte(`!@#$%^&*()`)},
			right:           [][]byte{[]byte(`!@#$%^&*()`), []byte(`{"foo": "bar"}`)},
			expectToBeEqual: true,
		},
		{
			left:            [][]byte{[]byte(`someBytes`)},
			right:           [][]byte{[]byte(`someBytes`), []byte(`moreBytes`)},
			expectToBeEqual: false,
		},
	}

	for _, v := range sut {
		areEqual := sakerhet.UnorderedEqualByteArrays(v.left, v.right)
		assert.Equal(suite.T(), v.expectToBeEqual, areEqual)
	}
}
