package amazon

import (
	"testing"
	"os"
	"github.com/eriklarko/purple-shopper/purple-shopper/products"
)

func TestOnlyPutBuyableProductsOnChannel(t *testing.T) {
	os.Chdir("..")
	channel := make(chan *products.RankedProduct, 3)

	products := products.ToRankedProducts(
		"http://www.amazon.com/Purple-Foiled-Milk-Chocolate-Hearts/dp/B0089ZTB1W/ref=pd_rhf_dp_s_cp_10_VD2D?ie=UTF8&refRID=11APPFCTSE47ZXGYE1WP",
		"http://www.amazon.com/gp/product/B00KNWPC5S/ref=s9_simh_gw_p193_d0_i1?pf_rd_m=ATVPDKIKX0DER&pf_rd_s=center-3&pf_rd_r=0VYEZDNQ6NQ2D1AF7QWX&pf_rd_t=101&pf_rd_p=1688200422&pf_rd_i=507846",
		"http://www.amazon.com/Purple-Chocolate-Ms-Candy-Pound/dp/B004XRJBE2/ref=sr_1_26?ie=UTF8&qid=1411982332&sr=8-26&keywords=purple",
	)

	PutBuyableProductsOnChannel(products, channel)

	if len(channel) != 1 {
		t.Errorf("Wrong number of products in channel. Expected 1, was %d\n", len(channel))
		t.Fatal()
	}

	select {
	case product := <-channel:
		if product.Product.Urls.Url.String() != "http://www.amazon.com/Purple-Foiled-Milk-Chocolate-Hearts/dp/B0089ZTB1W/ref=pd_rhf_dp_s_cp_10_VD2D?ie=UTF8&refRID=11APPFCTSE47ZXGYE1WP" {
			t.Errorf("Wrong product was added to channel, was %s\n", product.Product.Urls.Url.String())
		}
	}
}


