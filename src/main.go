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


func filter(c <-chan *products.ProductUrls, filter func(*products.ProductUrls)bool) <-chan *products.ProductUrls {
	output := make(chan *products.ProductUrls);
	go func() {
		for candidate := range c {
			if (filter(candidate)) {
				output <- candidate;
			}
		}
		close(output);
	}()
	return output;
}

func main() {
	start := time.Now()
	ranker := ranker.RankProductBasedOnAmountOfPurpleInImage

	for {
		log.Println("==================== SEARCHING FOR PURPLES ====================")


		productUrls := amazon.FindProducts(0, 10);
		unboughtProductUrls := filter(productUrls, productBoughtBefore)
		downloadedImages := downloadImages(unboughtProductUrls)
		rankedProducts := rankProducts(ranker, downloadedImages)
		buyableProducts := filterNonBuyableProducts(rankedProducts)

		highestRankedProduct := findHighestRankedProduct(buyableProducts)
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

func downloadImages(toDownloadChannel <-chan *products.ProductUrls) <-chan *products.Product {
	outchan := make(chan *products.Product)
	go func() {
		for toDownload := range toDownloadChannel {
			//log.Printf("Going to download %s\n", toDownload.ImageUrl)
			imageFile, error := downloader.DownloadImage(toDownload.ImageUrl)
			if error == nil {
				//log.Printf("Downloaded image %s to %s\n", urls.ImageUrl, imageFile)
				product := &products.Product{toDownload, imageFile}
				outchan <- product
			} else {
				log.Printf("Unable to download image: %v\n", error)
			}
		}
		close(outchan)
		log.Println("Finished downloading images")
	}()
	return outchan;
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

func rankProducts(ranker Ranker, toAnalyzeChannel <-chan *products.Product) <-chan *products.RankedProduct {
	outchan := make(chan *products.RankedProduct)
	go func() {
		for toAnalyze := range toAnalyzeChannel {
			rankedProduct := ranker(toAnalyze)
			//log.Printf("Ranking %s got me %v \n", toAnalyze.Urls.Url, rankedProduct)
			if rankedProduct != nil {
				log.Printf("Found a purple product! %s\n", rankedProduct.Product.Urls.Url)
				outchan <- rankedProduct
			}
		}

		close(outchan)
		log.Println("Finised ranking all products")
	}()
	return outchan;
}

func filterNonBuyableProducts(analyzedChannel <-chan *products.RankedProduct) <-chan *products.RankedProduct {
	outchan := make(chan *products.RankedProduct)
	go func() {
		var buffer []*products.RankedProduct

		for toCheckForBuyability := range analyzedChannel {
			log.Printf("Added %s to buyability queue\n", toCheckForBuyability.Product.Urls.Url)
			buffer = append(buffer, toCheckForBuyability)

			if len(buffer) >= 40 {
				log.Printf("Checking buyability of %d products\n", len(buffer))
				numberOfBuyableProducts := amazon.PutBuyableProductsOnChannel(buffer, outchan)
				log.Printf("Found %d buyable products\n", numberOfBuyableProducts)
				buffer = buffer[:0]
			}
		}

		if len(buffer) > 0 {
			log.Printf("Checking buyability of %d products\n", len(buffer))
			numberOfBuyableProducts := amazon.PutBuyableProductsOnChannel(buffer, outchan)
			log.Printf("Found %d buyable products\n", numberOfBuyableProducts)
		}

		close(outchan)
		log.Println("Finised filtering non-buyable products")
	}()

	return outchan;
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
