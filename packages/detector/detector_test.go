package detector

import (
	"testing"
	"log"
)


func TestSimilarity(t *testing.T) {
	sim, err := DetectSimilarity("131b4f18b60246cf6e0b99f0423d1bdd", "226ab72158bd99bc2f967b0a7f35e5ca")
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Comparing two different Woody Allen's photos, similarity is", sim)
}