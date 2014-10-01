package amazon

import (
	"log"
	"os"
	"github.com/eriklarko/purple-shopper/purple-shopper/products"
	"os/exec"
	"strings"
	"fmt"
)

func BuyProducts(products []*products.Product) {
	var args []string
	for _, product := range products {
		args = append(args, product.Urls.Url.String())
	}

	cmd := buildCasperScriptCommand("amazon/buyer/casperbuyer.js", args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	error := cmd.Run()
	if error != nil {
		log.Println("Unable to buy products :(")
		log.Fatal(error)
	}
}

func PutBuyableProductsOnChannel(products []*products.RankedProduct, c chan<- *products.RankedProduct) int {
	var urls []string
	urlToProductMap := make(map[string]int)
	for i, product := range products {
		urls = append(urls, product.Product.Urls.Url.String())
		urlToProductMap[product.Product.Urls.Url.String()] = i
	}

	cmd := buildCasperScriptCommand("amazon/buyer/items-can-be-bought.js", urls)
	cmd.Stderr = os.Stderr
	rawOutput, error := cmd.Output()
	if error != nil {
		log.Printf("Failed to check %d products for buyability, %v\n", len(products), error)
	}

	unprocessedOutput := string(rawOutput);
	lines := strings.Split(unprocessedOutput, "\n")
	numberOfBuyableProducts := 0
	for _, line := range lines {
		lineParts := strings.Split(line, ";")
		if len(lineParts) == 2 && lineParts[1] == "0" {

			productIndex, found := urlToProductMap[lineParts[0]]
			if found {
				numberOfBuyableProducts++
				log.Printf("%s was buyable!\n", products[productIndex].Product.Urls.Url)
				c <- products[productIndex]
			}
		}
	}

	return numberOfBuyableProducts
}

func buildCasperScriptCommand(script string, args []string) *exec.Cmd {
	phantomPath := "amazon/buyer/phantomjs/bin"
	if !strings.Contains(os.Getenv("PATH"), phantomPath) {
		path := fmt.Sprintf("%s:%s", os.Getenv("PATH"), phantomPath)
		os.Setenv("PATH", path)
	}
	casperPath := "amazon/buyer/casperjs/bin"
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
