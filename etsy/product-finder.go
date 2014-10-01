package etsy

import (
	"github.com/eriklarko/purple-shopper/purple-shopper/products"
	"bytes"
	"fmt"
	"github.com/eriklarko/purple-shopper/purple-shopper/randomkeyword"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"log"
	"strconv"
)

type Response struct {
	Count int
	Results []Result
	Pagination Pagination
}

type Result struct {
	Title string
	Price string
	Url string
	MainImage Image
	ShippingInfo []ShippingInfo
}

type Image struct {
	Url_170x135 string
}

type ShippingInfo struct {
	Destination_country_name string
	Primary_cost string
	Secondary_cost string
}

type Pagination struct {
	Next_offset *int
}

var limit int = 100
var apiKey *string = nil

func FindProducts(lowPrice, highPrice float64, c chan<- *products.ProductUrls, filter func(*products.ProductUrls)bool) {
	apiKey := getApiKey()
	keyword := randomkeyword.GenerateRandomSearchString()
	baseUrl := buildBaseUrl(lowPrice, highPrice, keyword, apiKey)

	offset := 0
	for {
		url := fmt.Sprintf("%s&offset=%d", baseUrl, offset)
		products, moreProducts := findProductsOnUrl(url)
		for _, product := range products {
			if filter(product) {
				c <-product
			}
		}

		if moreProducts {
			offset += limit
		} else {
			break
		}
	}

	log.Println("Closing channel")
	close(c)
}

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

func findProductsOnUrl(url string) ([]*products.ProductUrls, bool) {
	log.Println("Requesting " + url)
	var toReturn []*products.ProductUrls
	thereAreMoreProducts := false

	resp, error := http.Get(url)
	if(error == nil){
		body, error := ioutil.ReadAll(resp.Body)
		if(error == nil) {
			parsedJson := parseAsJson(body)
			for _, result := range parsedJson.Results {
				productUrl := result.Url
				imageUrl := result.MainImage.Url_170x135
				toReturn = append(toReturn, products.ToOneProductUrls(productUrl, imageUrl))
			}

			if parsedJson.Pagination.Next_offset != nil {
				thereAreMoreProducts = true
			}
		} else {
			log.Printf("Failed to read response from %s, %v\n", url, error)
		}

	} else {
		log.Printf("Unable to access %s, %v\n", url, error)
	}
	return toReturn, thereAreMoreProducts
}

func buildBaseUrl(lowPrice, highPrice float64, keyword, apiKey string) string {
	urlBuilder := bytes.NewBufferString("")
	urlBuilder.WriteString("https://openapi.etsy.com/v2")
	urlBuilder.WriteString("/listings/active")
	urlBuilder.WriteString("?fields=listing_id,title,price,url")
	urlBuilder.WriteString("&min_price=%f")
	urlBuilder.WriteString("&max_price=%f")
	urlBuilder.WriteString("&keywords=%s")
	urlBuilder.WriteString("&includes=")
	urlBuilder.WriteString("MainImage(url_170x135)")
		urlBuilder.WriteString(",ShippingInfo(destination_country_name,primary_cost,secondary_cost)")
		urlBuilder.WriteString("&api_key=%s")
	urlBuilder.WriteString("&limit=" + strconv.Itoa(limit))

	return fmt.Sprintf(urlBuilder.String(), lowPrice, highPrice, keyword, apiKey)
}

func parseAsJson(responseBody []byte) Response {
	var data Response
	error := json.Unmarshal(responseBody, &data)
	if error != nil {
		log.Printf("Failed to parse JSON response from >%s<, %v\n", string(responseBody), error)
	}
	return data;
}
