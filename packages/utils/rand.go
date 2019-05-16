package utils

import (
	"math/rand"

	"github.com/AplaProject/go-apla/packages/crypto"
)

type Rand struct {
	src *rand.Rand
}

func (r *Rand) BytesSeed(b []byte) *rand.Rand {
	seed, _ := crypto.CalcChecksum(b)
	r.src.Seed(int64(seed))
	return r.src
}

func NewRand(seed int64) *Rand {
	return &Rand{
		src: rand.New(rand.NewSource(seed)),
	}
}
