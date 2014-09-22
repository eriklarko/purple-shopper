package main

import (
  "log"
  "net/url"
  "errors"
  "strings"
  "math/rand"
  "time"
  "fmt"

  "github.com/PuerkitoBio/goquery"
)

type ProductUrls struct {
	Url *url.URL
	ImageUrl *url.URL
}

func findProductsOnRandomSearchPage(lowPrice, highPrice float64, c chan<- *ProductUrls) {
	keyword := generateRandomSearchString()
	category := getRandomCategory()

	baseUrl := fmt.Sprintf("http://www.amazon.com/s/search-alias%%3D%s&field-keywords=%s&low-price=%.2f&high-price=%.2f", category, keyword, lowPrice, highPrice)
	log.Println("Doing an Amazon search for " + keyword + " in category " + category + ": " + baseUrl)

	page := 0
	numberOfProductsFound := 0
	lastLoggedNumberOfProducts := 0
	shouldContinue := true
	for shouldContinue {
		page = page + 1

		url := fmt.Sprintf("%s&page=%d", baseUrl, page)
		numberOfNewProducts := findProductsOnSearchPageUrl(url, c)
		shouldContinue = numberOfNewProducts >= 10


		numberOfProductsFound += numberOfNewProducts
		if numberOfProductsFound - lastLoggedNumberOfProducts >= 150 {
			log.Printf("Found %d products\n", numberOfProductsFound)
			lastLoggedNumberOfProducts = numberOfProductsFound
		}
	}

	log.Printf("Found a total of %d products\n", numberOfProductsFound)
	c <- nil
}

func getRandomCategory() string {
	categories := []string {
		"appliances",
		"arts-crafts",
		"automotive",
		"baby-products",
		"beauty",
		"popular",
		"mobile",
		"collectibles",
		"computers",
		"electronics",
		"grocery",
		"hpc",
		"garden",
		"industrial",
		"fashion-luggage",
		"magazines",
		"mi",
		"office-products",
		"lawngarden",
		"pets",
		"pantry",
		"software",
		"sporting",
		"tools",
		"toys-and-games",
		"wine",
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return strings.ToLower(categories[r.Intn(len(categories))])
}

func findProductsOnSearchPageUrl(url string, c chan<- *ProductUrls) int {
	doc, error := goquery.NewDocument(url)
	if error != nil {
		log.Fatal(error)
	}

	return findProductsOnSearchPage(doc, c)
}

func findProductsOnSearchPage(doc *goquery.Document, c chan<- *ProductUrls) int {
	numberOfProductsFound := 0
	doc.Find(".productImage").Each(func (i int, image *goquery.Selection) {
		product, error := extractProduct(image);
		if error == nil {
			numberOfProductsFound++
			c <- &product
		} else {
			log.Printf("Unable to extract product from goquery selection: %v\n", error)
		}
	})

	return numberOfProductsFound
}

func extractProduct(s *goquery.Selection) (ProductUrls, error) {
	parentWithLink := findParentWithHref(s)
	if parentWithLink == nil {
		return ProductUrls{}, errors.New("No parent with href attribute found")
	}

	imageUrl, imageUrlError := attrToUrl(s, "src")
	if imageUrlError == nil {
		link,_ := attrToUrl(parentWithLink, "href")
		return ProductUrls {link, imageUrl}, nil
	}

	return ProductUrls{}, imageUrlError
}

func findParentWithHref(s *goquery.Selection) *goquery.Selection {
	// TODO: Do loop to find first parent with link
	return s.Parent().Parent()
}

func attrToUrl(s *goquery.Selection, attr string) (*url.URL, error) {
	link, exists := s.Attr(attr);
	if exists {
		return url.Parse(link)
	}

	return nil, errors.New("Attr " + attr + " not found")
}
