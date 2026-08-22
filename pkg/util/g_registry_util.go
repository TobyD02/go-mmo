package util

import (
	"fmt"
	"math/rand/v2"
)

func GetRandomIDFromRegistry[T any](registry map[string]T) (string, error) {
	if len(registry) == 0 {
		return "", fmt.Errorf("registry is empty")
	}
	target := rand.IntN(len(registry))

	count := 0
	for id := range registry {
		if count == target {
			return id, nil
		}

		count++
	}

	return "", fmt.Errorf("Something went wrong in GetRandomIDFromRegistry?!")
}
