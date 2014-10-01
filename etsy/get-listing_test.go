package etsy

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/eriklarko/purple-shopper/purple-shopper/products"
	"os"
	"log"
)

func TestDoListingRequest(t *testing.T) {
	os.Chdir("..")

	c := make(chan *products.ProductUrls, 1000000)
	FindProducts(0.0, 0.0, c, dummyFilter)

	productsFound := 0
	for _ = range c {
		productsFound++
	}

	log.Printf("Found %d products\n", productsFound)
	assert.True(t, productsFound > 0)
}
func dummyFilter(product *products.ProductUrls) bool {
	return true
}
