package main

/*
   TODO: Don't buy clothes
   TODO: Don't buy books
   TODO: Don't buy apps
   TODO: Don't buy sex toys
*/
import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Product struct {
	Urls  *ProductUrls
	Image string
}

type Ranker func ([]*Product, *Product, *os.File) int

func main() {
	ranker := RankProductBasedOnAmountOfPurpleInImage

	// TODO: Make multiple parallel calls to the findRandom thingie
	// TODO: Limit products on price and availability
	var productUrls []*ProductUrls
	for i := 0; i < 5; i++ {
		productUrls = append(productUrls, findProductsOnRandomSearchpage()...)
	}
	if len(productUrls) == 0 {
		log.Fatal("No products found!")
	}

	products := downloadImages(productUrls)
	defer cleanUp(products)
	if len(products) == 0 {
		log.Fatal("No images downloaded!")
	}

	highestRankedProduct := findHighestRankedProduct(ranker, products)
	if highestRankedProduct == nil {
		fmt.Println("Did not find a good enough product :(")
	} else {
		buyProduct(highestRankedProduct)
	}
}

func downloadImages(productUrls []*ProductUrls) []*Product {
	log.Printf("Downloading %d images...\n", len(productUrls))
	var products []*Product
	for i, urls := range productUrls {
		log.Printf("Downloading image %d\n", i+1)
		imageFile, error := downloadImage(urls.ImageUrl)
		if error == nil {
			products = append(products, &Product{urls, imageFile})
		} else {
			log.Println(error)
		}
	}
	log.Println("... done")
	return products
}

func downloadImage(url *url.URL) (string, error) {
	res, err := http.Get(url.String())
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", err
	}

	file, err := os.Create("image_" + getImageName(url))
	defer file.Close()
	if err != nil {
		return "", err
	}

	_, err = file.Write(data)
	if err != nil {
		return "", err
	}
	file.Sync()
	//log.Printf("Downloaded %s to %s\n", url.String(), file.Name())
	return file.Name(), nil
}

func getImageName(url *url.URL) string {
	urlString := url.String()
	slashIndex := strings.LastIndex(urlString, "/")

	return urlString[slashIndex + 1:]
}

func findHighestRankedProduct(ranker Ranker, products []*Product) *Product {
	log.Printf("Analyzing %d images...\n", len(products))
	highestRank := 0
	var highestRankedProduct *Product = nil
	for i, product := range products {
		log.Printf("Analyzing image %d\n", i+1)

		imageFile, error := os.Open(product.Image)
		if error != nil {
			imageFile.Close()
			log.Printf("Unable to open file for %v. %v\n", product.Urls.Url, error)
			continue
		}

		productRank := ranker(products, product, imageFile)
		imageFile.Close()
		if productRank > highestRank {
			highestRank = productRank
			highestRankedProduct = product
		}
	}

	if highestRankedProduct != nil {
		log.Printf("I found %v which ranked at %d!", highestRankedProduct.Urls.Url, highestRank)
	}

	log.Println("... done!")
	return highestRankedProduct
}

func buyProduct(product *Product) {
	fmt.Printf("I AM GONNA BUY %v\n", product.Urls.Url)
}

func cleanUp(products []*Product) {
	for _, product := range products {
		os.Remove(product.Image)
	}
	log.Printf("Removed %d images\n", len(products))
}
