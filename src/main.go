package main

import (
	"log"
	"os"
	"time"
	"randomkeyword"
	"amazon"
	"products"
	"downloader"
	"ranker"
)

type Ranker func (*products.Product) *products.RankedProduct

var boughtProducts []string = nil

func main() {
	start := time.Now()
	ranker := ranker.RankProductBasedOnAmountOfPurpleInImage

	for {
		log.Println("==================== SEARCHING FOR PURPLES ====================")

		toDownloadChannel := make(chan *products.ProductUrls, 200)
		toAnalyzeChannel := make(chan *products.Product, 70)
		analyzedChannel := make(chan *products.RankedProduct, 10000)
		buyableChannel := make(chan *products.RankedProduct, 10000)

		go amazon.FindProducts(0, 100, toDownloadChannel, productBoughtBefore)
		go downloadImages(toDownloadChannel, toAnalyzeChannel)
		go rankProducts(ranker, toAnalyzeChannel, analyzedChannel)
		go filterNonBuyableProducts(analyzedChannel, buyableChannel)

		highestRankedProduct := findHighestRankedProduct(buyableChannel)
		if highestRankedProduct == nil {
			log.Println("Did not find a good enough product :(")
		} else {
			var products []*products.Product
			products = append(products, highestRankedProduct)
			amazon.BuyProducts(products)
			os.Exit(0)
		}
	}
	elapsed := time.Since(start)
	log.Printf("Running time %s", elapsed)
}

func downloadImages(toDownloadChannel chan *products.ProductUrls, toAnalyzeChannel chan<- *products.Product) {
	for toDownload := range toDownloadChannel {
		//log.Printf("Going to download %s\n", toDownload.ImageUrl)
		imageFile, error := downloader.DownloadImage(toDownload.ImageUrl)
		if error == nil {
			//log.Printf("Downloaded image %s to %s\n", urls.ImageUrl, imageFile)
			product := &products.Product{toDownload, imageFile}
			toAnalyzeChannel <- product
		} else {
			log.Printf("Unable to download image: %v\n", error)
		}
	}

	close(toAnalyzeChannel)
	log.Println("Finished downloading images")
}

func productBoughtBefore(urls *products.ProductUrls) bool {
	if boughtProducts == nil {
		lines, error := randomkeyword.ReadLines("bought-products.txt")
		if error == nil {
			boughtProducts = lines
		} else {
			log.Fatalf("Unable to open bought products log: %v\n", error)
		}
	}

	return !stringInSlice(urls.Url.String(), boughtProducts);
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func rankProducts(ranker Ranker, toAnalyzeChannel <-chan *products.Product, analyzedChannel chan<- *products.RankedProduct) {
	for toAnalyze := range toAnalyzeChannel {
		rankedProduct := ranker(toAnalyze)
		//log.Printf("Ranking %s got me %v \n", toAnalyze.Urls.Url, rankedProduct)
		if rankedProduct != nil {
			log.Printf("Found a purple product! %s\n", rankedProduct.Product.Urls.Url)
			analyzedChannel <- rankedProduct
		}
	}

	close(analyzedChannel)
	log.Println("Finised ranking all products")
}

func filterNonBuyableProducts(analyzedChannel <-chan *products.RankedProduct, buyableChannel chan<- *products.RankedProduct) {
	var buffer []*products.RankedProduct

	for toCheckForBuyability := range analyzedChannel {
		log.Printf("Added %s to buyability queue\n", toCheckForBuyability.Product.Urls.Url)
		buffer = append(buffer, toCheckForBuyability)

		if len(buffer) >= 40 {
			log.Printf("Checking buyability of %d products\n", len(buffer))
			numberOfBuyableProducts := amazon.PutBuyableProductsOnChannel(buffer, buyableChannel)
			log.Printf("Found %d buyable products\n", numberOfBuyableProducts)
			buffer = buffer[:0]
		}
	}

	if len(buffer) > 0 {
		log.Printf("Checking buyability of %d products\n", len(buffer))
		numberOfBuyableProducts := amazon.PutBuyableProductsOnChannel(buffer, buyableChannel)
		log.Printf("Found %d buyable products\n", numberOfBuyableProducts)
	}

	close(buyableChannel)
	log.Println("Finised filtering non-buyable products")
}

func findHighestRankedProduct(buyableChannel <-chan *products.RankedProduct) *products.Product {
	highestRank := 0
	var highestRankedProduct *products.Product = nil

	for buyableRankedProduct := range buyableChannel {
		if buyableRankedProduct.Rank > highestRank {
			highestRank = buyableRankedProduct.Rank
			highestRankedProduct = buyableRankedProduct.Product

			log.Printf("Found new top product! %v ranked at %d\n", highestRankedProduct.Urls.Url, highestRank)
		}
	}
	log.Println("No more rankings to process")

	if highestRankedProduct != nil {
		log.Printf("I found %v which ranked at %d!", highestRankedProduct.Urls.Url, highestRank)
	}

	return highestRankedProduct
}

func cleanUp(products []*products.Product) {
	for _, product := range products {
		os.Remove(product.Image)
	}
	log.Printf("Removed %d images\n", len(products))
}
