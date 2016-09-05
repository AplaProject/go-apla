package test

import (
	"time"
	"math/rand"
)

// Generates a random []bytes.
func RandBytes(length int) ([]byte, int64) {
	seed := time.Now().UTC().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	ret := make([]byte, length)
	for ; length > 0; length-- {
		ret[length-1] = byte(rng.Intn(256))
	}
	return ret, seed
}
