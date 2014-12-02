package etsy

import (
	"net/http"
	"log"
	"io/ioutil"
	"fmt"
)

func AddProductToCart(listingId int) {
	// Skicka till /users/__SELF__/carts&listing_id=
	resp, error := http.Get(apiUrl + fmt.Sprintf("/users/__SELF__/carts&listing_id=%d&apikey=%s", listingId, getApiKey()))
	if error != nil {
		log.Panicf("Unable to add product to cart. Failed on HTTP request: %v\n", error)
	}

	/*body*/_, error := ioutil.ReadAll(resp.Body)
	if error != nil {
		log.Panicf("Unable to add product to cart. Failed reading HTTP response: %v\n", error)
	}

	// TODO: Parse response for indication that the item was successfully added to cart
}

func EmptyAllCarts() {
	cartIds, error := getAllCartIds();
	if error != nil {
		log.Panicf("Unable to get cartIds: %v\n", error)
	}

	for _, cartId := range cartIds {
		emptyCart(cartId)
	}
}

func getAllCartIds() ([]int, error) {
	resp, error := http.Get(apiUrl + "/users/__SELF__/carts&apikey=" + getApiKey())
	if error != nil {
		return nil, fmt.Errorf("Error while requesting all carts: %v\n", error)
	}

	body, error := ioutil.ReadAll(resp.Body)
	if error != nil {
		return nil, fmt.Errorf("Error while reading response from all carts: %v\n", error)
	}

	cartIds, error := parseCartsJson(body)
	if error != nil {
		return nil, fmt.Errorf("Error while parsing cart ID:s: %v\n", error)
	}

	return cartIds, nil
}

func parseCartsJson(responseBody []byte) ([]int, error) {
	var ids []int

	return ids, nil;
}

func emptyCart(cartId int) error {
	url := apiUrl + fmt.Sprintf("/users/__SELF__/carts/%d&apikey=%s", cartId, getApiKey())
	resp, error := http.NewRequest("DELETE", url, nil)
	if error != nil {
		return fmt.Errorf("Unable to empty cart %d. Error with HTTP request: %v\n", cartId, error)
	}

	/*body*/_, error := ioutil.ReadAll(resp.body)
	if error != nil {
		return fmt.Errorf("Unable to empty cart %d. Error while reading response: %v\n", cartId, error)
	}

	// TODO: Parse response for indication that the cart was removed

	return nil
}
