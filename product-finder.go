package main

import (
  "fmt"
  "log"
  "net/url"
  "errors"

  "github.com/PuerkitoBio/goquery"
)

type ProductUrls struct {
	Url *url.URL
	ImageUrl *url.URL
}

func findProductsOnRandomSearchpage() []*ProductUrls {
	keyword := getRandomSearchKeyword()

	// TODO: This URL only show books :)
	// TODO: Use http://www.amazon.com/s/search-alias%3D[CATEGORY]&field-keywords=[KEYWORDS]
	url := "http://www.amazon.com/s/field-keywords=" + keyword;
	fmt.Println("Finding products on " + url)
	doc, error := goquery.NewDocument(url)
	if error != nil {
		log.Fatal(error)
	}

	return findProductsOnSearchPage(doc)
}

func getRandomSearchKeyword() string {
	return generateRandomSearchString()
}

func findProductsOnSearchPage(doc *goquery.Document) []*ProductUrls {
	var products []*ProductUrls
	doc.Find(".productImage").Each(func (i int, image *goquery.Selection) {
		product, error := extractProduct(image);
		if error == nil {
			products = append(products, &product)
		} else {
			log.Println(error)
		}
	})
	return products
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