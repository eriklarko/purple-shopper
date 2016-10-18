package amazon

// TODO: Rename

import (
  "log"
  "net/url"
  "errors"
  "strings"
  "math/rand"
  "time"
  "fmt"

  "github.com/PuerkitoBio/goquery"
  "randomkeyword"
  "products"
)

func FindProducts() <-chan *products.ProductUrls {
	outchan := make(chan *products.ProductUrls)
	go func() {
		keyword := randomkeyword.GenerateRandomSearchString()
		category := getRandomCategory()

		baseUrl := "https://www.amazon.com/s/ref=nb_sb_noss_2?url=search-alias%3D" +category+ "&field-keywords=" + keyword
		log.Println("Doing an Amazon search for " + keyword + " in category " + category + ": " + baseUrl)

		page := 0
		numberOfProductsFound := 0
		lastLoggedNumberOfProducts := 0
		shouldContinue := true
		for shouldContinue {
			page = page + 1

			url := fmt.Sprintf("%s&page=%d", baseUrl, page)
			numberOfNewProducts := findProductsOnSearchPageUrl(url, outchan)
			log.Printf("Found %d products on %s\n", numberOfNewProducts, url)
			shouldContinue = numberOfNewProducts >= 10

			numberOfProductsFound += numberOfNewProducts
			if numberOfProductsFound - lastLoggedNumberOfProducts >= 150 {
				log.Printf("Found %d products\n", numberOfProductsFound)
				lastLoggedNumberOfProducts = numberOfProductsFound
			}
		}

		log.Printf("Found a total of %d products\n", numberOfProductsFound)
		close(outchan)
	}()

	return outchan;
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

func findProductsOnSearchPageUrl(url string, c chan<- *products.ProductUrls) int {
	doc, error := goquery.NewDocument(url)
	if error != nil {
		log.Printf("Unable to build goquery document from url %s: %v\n", url, error)
		return 0
	}

	return findProductsOnSearchPage(doc, c)
}

func findProductsOnSearchPage(doc *goquery.Document, c chan<- *products.ProductUrls) int {
	numberOfProductsFound := 0
	doc.Find(".s-access-image").Each(func (i int, image *goquery.Selection) {
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

func extractProduct(s *goquery.Selection) (products.ProductUrls, error) {
	parentWithLink := findParentWithHref(s)
	if parentWithLink == nil {
		return products.ProductUrls{}, errors.New("No parent with href attribute found")
	}

	imageUrl, imageUrlError := attrToUrl(s, "src")
	if imageUrlError == nil {
		link,_ := attrToUrl(parentWithLink, "href")
		return products.ProductUrls {link, imageUrl}, nil
	}

	return products.ProductUrls{}, imageUrlError
}

func findParentWithHref(s *goquery.Selection) *goquery.Selection {
	// TODO: Do loop to find first parent with link
	return s.Parent()
}

func attrToUrl(s *goquery.Selection, attr string) (*url.URL, error) {
	link, exists := s.Attr(attr);
	if exists {
		return url.Parse(link)
	}

	return nil, errors.New("Attr " + attr + " not found")
}
