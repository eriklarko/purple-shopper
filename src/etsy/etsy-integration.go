package etsy

import (
	"github.com/eriklarko/purple-shopper/purple-shopper/randomkeyword"
	"log"
)

var apiKey *string = nil
var apiUrl string = "https://openapi.etsy.com/v2"

func getApiKey() string {
	if apiKey == nil {
		lines, error := randomkeyword.ReadLines("etsy/apikey")
		if error != nil {
			log.Fatalf("Unable to read Etsy API Key: %v\n", error)
		}
		if len(lines) == 0 {
			log.Fatalf("Etsy API Key file was empty\n")
		}
		apiKey = &lines[0]
	}
	return *apiKey
}

