package abstractedcontainers

import "fmt"

func UnorderedEqualByteArrays(first, second [][]byte) bool {
	var firstA []any
	var secondA []any

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
