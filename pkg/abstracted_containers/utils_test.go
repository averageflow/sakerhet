//go:build !integration

package abstractedcontainers_test

import (
	"testing"

	abstractedcontainers "github.com/averageflow/sakerhet/pkg/abstracted_containers"
	"github.com/stretchr/testify/assert"
)

func TestUnorderedEqual(t *testing.T) {
	type unorderedEqualTestCase struct {
		left            []any
		right           []any
		expectToBeEqual bool
	}

	sut := []unorderedEqualTestCase{
		{left: []any{"one", "two"}, right: []any{"two", "one"}, expectToBeEqual: true},
		{left: []any{true, true, false}, right: []any{false, true, true}, expectToBeEqual: true},
		{left: []any{1, 2, 3}, right: []any{3, 2, 1}, expectToBeEqual: true},
		{left: []any{1, 3}, right: []any{2, 3}, expectToBeEqual: false},
		{left: []any{"one"}, right: []any{"two"}, expectToBeEqual: false},
	}

	for _, v := range sut {
		areEqual := abstractedcontainers.UnorderedEqual(v.left, v.right)
		assert.Equal(t, v.expectToBeEqual, areEqual)
	}
}

func TestUnorderedEqualByteArrays(t *testing.T) {
	type unorderedEqualTestCase struct {
		left            [][]byte
		right           [][]byte
		expectToBeEqual bool
	}

	sut := []unorderedEqualTestCase{
		{left: [][]byte{[]byte(`{"foo": "bar"}`), []byte(`!@#$%^&*()`)}, right: [][]byte{[]byte(`!@#$%^&*()`), []byte(`{"foo": "bar"}`)}, expectToBeEqual: true},
		{left: [][]byte{[]byte(`someBytes`)}, right: [][]byte{[]byte(`someBytes`), []byte(`moreBytes`)}, expectToBeEqual: false},
	}

	for _, v := range sut {
		areEqual := abstractedcontainers.UnorderedEqualByteArrays(v.left, v.right)
		assert.Equal(t, v.expectToBeEqual, areEqual)
	}
}
