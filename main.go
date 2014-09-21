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

func main() {
	start := time.Now()
	log.Println("Starting concurrent Purple Shopper")
	ranker := RankProductBasedOnAmountOfPurpleInImage

	toDownloadChannel := make(chan *ProductUrls, 10000)
	toAnalyzeChannel := make(chan *Product, 10000)
	analyzedChannel := make(chan *RankedProduct, 10000)


	for {
		go findProductsOnRandomSearchPage(toDownloadChannel)
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
	moarMessages := true
	for moarMessages {
		select {
		case toDownload := <-toDownloadChannel:
			if toDownload == nil {
				toAnalyzeChannel <- nil
				log.Println("Finished downloading images")
				moarMessages = false
				break
			}

			// TODO: This should be in it's on goroutine
			product, error := downloadProductImage(toDownload)
			//log.Printf("Downloaded %v\n", product.Urls.ImageUrl)
			if error == nil {
				toAnalyzeChannel <- product
			}
		}
	}
}

func downloadProductImage(urls *ProductUrls) (*Product, error) {
	imageFile, error := downloadImage(urls.ImageUrl)
	if error == nil {
		return &Product{urls, imageFile}, nil
	} else {
		log.Println(error)
		return nil, error
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
	//log.Printf("Downloaded %s to %s\n", url.String(), file.Name())
	return file.Name(), nil
}

func getImageName(url *url.URL) string {
	urlString := url.String()
	slashIndex := strings.LastIndex(urlString, "/")

	return urlString[slashIndex + 1:]
}

func rankProducts(ranker Ranker, toAnalyzeChannel <-chan *Product, analyzedChannel chan<- *RankedProduct) {
	moarMessages := true
	for moarMessages {
		select {
		case toAnalyze := <-toAnalyzeChannel:
			if toAnalyze == nil {
				analyzedChannel <- nil
				log.Println("Finised ranking all products")
				moarMessages = false
				break
			}

			go ranker(toAnalyze, analyzedChannel)
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
