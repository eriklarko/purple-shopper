package etsy

func AddProductToCart(listingId int) {
	// Skicka till /users/__SELF__/carts
}

func EmptyAllCarts() {
	// DELETE till /users/__SELF__/carts/:cart_id
	// :cart_id från GET /users/__SELF__/carts
}
