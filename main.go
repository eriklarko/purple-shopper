package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"os/exec"
	"fmt"
)

type Product struct {
	Urls  *ProductUrls
	Image string
}

type RankedProduct struct {
	product *Product
	rank int
}

type Ranker func (*Product, chan<- *RankedProduct)

var boughtProducts []string = nil

func main() {
	start := time.Now()
	log.Println("Starting concurrent Purple Shopper")
	ranker := RankProductBasedOnAmountOfPurpleInImage

	toDownloadChannel := make(chan *ProductUrls, 10000)
	toAnalyzeChannel := make(chan *Product, 100)
	analyzedChannel := make(chan *RankedProduct, 10000)
	buyableChannel := make(chan *RankedProduct, 10000)

	for {
		go findProductsOnRandomSearchPage(0, 100, toDownloadChannel)
		go downloadImages(toDownloadChannel, toAnalyzeChannel)
		go rankProducts(ranker, toAnalyzeChannel, analyzedChannel)
		go filterNonBuyableProducts(analyzedChannel, buyableChannel)

		highestRankedProduct := findHighestRankedProduct(buyableChannel)
		if highestRankedProduct == nil {
			log.Println("Did not find a good enough product :(")
		} else {
			var products []*Product
			products = append(products, highestRankedProduct)
			buyProducts(products)
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

func filterNonBuyableProducts(analyzedChannel <-chan *RankedProduct, buyableChannel chan<- *RankedProduct) {
	var buffer []*RankedProduct

	moarMessages := true
	for moarMessages {
		select {
		case toCheckForBuyability := <-analyzedChannel:
			if toCheckForBuyability == nil {
				buyableChannel <- nil
				log.Println("Finised filtering non-buyable products")
				moarMessages = false
			} else {
				buffer = append(buffer, toCheckForBuyability)
			}

			if len(buffer) >= 40 || !moarMessages {
				log.Printf("Checking buyability of %d products\n", len(buffer))
				numberOfBuyableProducts := putBuyableProductsOnChannel(buffer, buyableChannel)
				log.Printf("Found a total of %d products\n", numberOfBuyableProducts)
			}
		}
	}
}

func putBuyableProductsOnChannel(products []*RankedProduct, c chan<- *RankedProduct) int {
	var urls []string
	urlToProductMap := make(map[string]*RankedProduct)
	for _, product := range products {
		urls = append(urls, product.product.Urls.Url.String())
		urlToProductMap[product.product.Urls.Url.String()] = product
	}

	cmd := BuildCasperScriptCommand("buyer/items-can-be-bought.js", urls)
	rawOutput, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to check %d products for buyability, %v\n", len(products), err)
	}

	unprocessedOutput := string(rawOutput);
	lines := strings.Split(unprocessedOutput, "\n")
	numberOfBuyableProducts := 0
	for _, line := range lines {
		lineParts := strings.Split(line, ";")
		if len(lineParts) == 2 && lineParts[1] == "0" {

			product, found := urlToProductMap[lineParts[0]]
			if found {
				numberOfBuyableProducts++
				c <- product
			}
		}
	}

	return numberOfBuyableProducts
}

func findHighestRankedProduct(buyableChannel <-chan *RankedProduct) *Product {
	highestRank := 0
	var highestRankedProduct *Product = nil

	moarMessages := true
	for moarMessages {
		select {
		case buyableRankedProduct := <-buyableChannel:
			if buyableRankedProduct == nil {
				log.Println("No more rankings to process")
				moarMessages = false
				break
			}

			if buyableRankedProduct.rank > highestRank {
				highestRank = buyableRankedProduct.rank
				highestRankedProduct = buyableRankedProduct.product

				log.Printf("Found new top product! %v ranked at %d\n", highestRankedProduct.Urls.Url, highestRank)
			}
		}
	}

	if highestRankedProduct != nil {
		log.Printf("I found %v which ranked at %d!", highestRankedProduct.Urls.Url, highestRank)
	}

	return highestRankedProduct
}

func buyProducts(products []*Product) {
	var args []string
	for _, product := range products {
		args = append(args, product.Urls.Url.String())
	}

	cmd := BuildCasperScriptCommand("buyer/casperbuyer.js", args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Println("Unable to buy products :(")
		log.Fatal(err)
	}
}

func BuildCasperScriptCommand(script string, args []string) *exec.Cmd {
	phantomPath := "buyer/phantomjs/bin"
	if !strings.Contains(os.Getenv("PATH"), phantomPath) {
		path := fmt.Sprintf("%s:%s", os.Getenv("PATH"), phantomPath)
		os.Setenv("PATH", path)
	}
	casperPath := "buyer/casperjs/bin"
	if !strings.Contains(os.Getenv("PATH"), casperPath) {
		path := fmt.Sprintf("%s:%s", os.Getenv("PATH"), casperPath)
		os.Setenv("PATH", path)
	}

	var realArgs []string;
	realArgs = append(realArgs, "casperjs");
	realArgs = append(realArgs, script);
	realArgs = append(realArgs, args...);

	cmd := exec.Command("casperjs");
	cmd.Args = realArgs

	return cmd
}

func cleanUp(products []*Product) {
	for _, product := range products {
		os.Remove(product.Image)
	}
	log.Printf("Removed %d images\n", len(products))
}
