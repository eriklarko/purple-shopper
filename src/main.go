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
	for {
		log.Println("==================== SEARCHING FOR PURPLES ====================")

		/* Make a random search on amazon.com and read the image and product details urls from the html
		 * e.g.
                 *   {
		 *	imageUrl: "https://images-na.ssl-images-amazon.com/images/I/5188yugyWZL._AC_US160_.jpg",
		 *	productUrl: "https://www.amazon.com/Dealzip-Fashion-Octopus-Cthulhu-Knitting/dp/B00VFX2NTW/ref=sr_1_1?ie=UTF8&qid=1476899795&sr=8-1&keywords=random"
		 *   }
		 */
		productUrls := amazon.FindProducts();

		// Download the image referenced by the imageUrl above
		downloadedImages := downloadImages(productUrls)

		/* Give the image a score based on how purple it is. Between 0 and 441 :)
		 * The object above is extended with the rank:
		 *  {
		 *      imageUrl: ...,
		 *	productUrl: ...,
		 *	rank: 410
		 *  }
		 */
		rankedProducts := rankProducts(
			ranker.RankProductBasedOnAmountOfPurpleInImage,
			downloadedImages,
		)

		// Throw away any products that don't ship to Sweden
		//buyableProducts := filterNonBuyableProducts(rankedProducts)

		// Throw away any products we've already bought
		unboughtBuyableProducts := filter(rankedProducts, productHasBeenBoughtBefore)

		// Find the unbought and buyable product with the highest purple-score
		highestRankedProduct := findHighestRankedProduct(unboughtBuyableProducts)

		if highestRankedProduct == nil {
			log.Println("Did not find a good enough product :( Will try again!")
		} else {
			// Buy the product!
			amazon.Buy(highestRankedProduct)
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
			imageFile, error := downloader.DownloadImage(toDownload.ImageUrl)
			if error == nil {
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

func filter(c <-chan *products.RankedProduct, filter func(*products.RankedProduct)bool) <-chan *products.RankedProduct {
	output := make(chan *products.RankedProduct);
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

func productHasBeenBoughtBefore(urls *products.RankedProduct) bool {
	if boughtProducts == nil {
		lines, error := randomkeyword.ReadLines("bought-products.txt")
		if error == nil {
			boughtProducts = lines
		} else {
			log.Fatalf("Unable to open bought products log: %v\n", error)
		}
	}

	return !stringInSlice(urls.Product.Urls.Url.String(), boughtProducts);
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
