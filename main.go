package main

/*
   TODO: Store which products have been bought so that we don't accidentally buy them again
   TODO: Don't suggest products that aren't available
   TODO: Don't suggest products where you have to eg. select size before the product can be added to the cart
*/
import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Product struct {
	Urls  *ProductUrls
	Image string
}

type RankedProduct struct {
	product *Product
	rank int
}

// Sadly, the ranker is responsible for closing and removing the file
type Ranker func (*Product, chan<- *RankedProduct)

var boughtProducts []string = nil

func main() {
	start := time.Now()
	log.Println("Starting concurrent Purple Shopper")
	ranker := RankProductBasedOnAmountOfPurpleInImage

	toDownloadChannel := make(chan *ProductUrls, 10000)
	toAnalyzeChannel := make(chan *Product, 100)
	analyzedChannel := make(chan *RankedProduct, 10000)


	for {
		go findProductsOnRandomSearchPage(0, 100, toDownloadChannel)
		go downloadImages(toDownloadChannel, toAnalyzeChannel)
		go rankProducts(ranker, toAnalyzeChannel, analyzedChannel)

		highestRankedProduct := findHighestRankedProduct(analyzedChannel)
		if highestRankedProduct == nil {
			log.Println("Did not find a good enough product :(")
		} else {
			buyProduct(highestRankedProduct)
			break
		}
	}

	elapsed := time.Since(start)
	log.Printf("Running time %s", elapsed)
}

func downloadImages(toDownloadChannel <-chan *ProductUrls, toAnalyzeChannel chan<- *Product) {
	select {
	case toDownload := <-toDownloadChannel:
		if toDownload == nil {
			toAnalyzeChannel <- nil
			log.Println("Finished downloading images")
		} else if !productBoughtBefore(toDownload){
			go downloadProductImage(toDownload, toAnalyzeChannel)
			downloadImages(toDownloadChannel, toAnalyzeChannel)
		}
	}
}

func productBoughtBefore(urls *ProductUrls) bool {
	if boughtProducts == nil {
		lines, error := ReadLines("bought-products.txt")
		if error == nil {
			boughtProducts = lines
		} else {
			log.Fatal(error)
		}
	}

	return stringInSlice(urls.Url.String(), boughtProducts);
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func downloadProductImage(urls *ProductUrls, toAnalyzeChannel chan<- *Product) {
	imageFile, error := downloadImage(urls.ImageUrl)
	if error == nil {
		product := &Product{urls, imageFile}
		toAnalyzeChannel <- product
	} else {
		log.Println(error)
	}
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
	return file.Name(), nil
}

func getImageName(url *url.URL) string {
	urlString := url.String()
	slashIndex := strings.LastIndex(urlString, "/")

	return urlString[slashIndex + 1:]
}

func rankProducts(ranker Ranker, toAnalyzeChannel <-chan *Product, analyzedChannel chan<- *RankedProduct) {
	select {
	case toAnalyze := <-toAnalyzeChannel:
		if toAnalyze == nil {
			analyzedChannel <- nil
			log.Println("Finised ranking all products")
		} else {
			go ranker(toAnalyze, analyzedChannel)
			rankProducts(ranker, toAnalyzeChannel, analyzedChannel)
		}
	}
}

func findHighestRankedProduct(analyzedChannel <-chan *RankedProduct) *Product {
	highestRank := 0
	var highestRankedProduct *Product = nil

	moarMessages := true
	for moarMessages {
		select {
		case rankedProduct := <-analyzedChannel:
			if rankedProduct == nil {
				log.Println("No more rankings to process")
				moarMessages = false
				break
			}

			if rankedProduct.rank > highestRank {
				highestRank = rankedProduct.rank
				highestRankedProduct = rankedProduct.product

				log.Printf("Found new top product! %v ranked at %d\n", highestRankedProduct.Urls.Url, highestRank)
			}
		}
	}

	if highestRankedProduct != nil {
		log.Printf("I found %v which ranked at %d!", highestRankedProduct.Urls.Url, highestRank)
	}

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
