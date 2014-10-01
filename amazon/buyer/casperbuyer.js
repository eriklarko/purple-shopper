// TODO: Don't have separate steps, just look for things like "Continue", "Proceed", "Place" etc and just press whatever's there.

var fs = require('fs');
var utils = require("utils");
var casper = require('casper').create();
casper.userAgent('Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)');

////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////

var amazonLoginPage = 'https://www.amazon.com/ap/signin?_encoding=UTF8&openid.assoc_handle=usflex&openid.claimed_id=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0%2Fidentifier_select&openid.identity=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0%2Fidentifier_select&openid.mode=checkid_setup&openid.ns=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0&openid.ns.pape=http%3A%2F%2Fspecs.openid.net%2Fextensions%2Fpape%2F1.0&openid.pape.max_auth_age=0&openid.return_to=https%3A%2F%2Fwww.amazon.com%2Fgp%2Fyourstore%2Fhome%3Fie%3DUTF8%26ref_%3Dnav_custrec_signin';
var amazonCartUrl = "http://www.amazon.com/gp/cart/view.html";
var productUrls =  casper.cli.args;

////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////

function login(username, password) {
	this.capture("amazon/images/login-before.png");
	fs.write("amazon/dumps/login-before.html", this.getHTML(), 'w');

	this.thenEvaluate(function (username, password) {
		// TODO: if auth-email exists, do this
		document.getElementById('auth-email').value = username;
		document.getElementById('auth-password').value = password;
		document.forms["signIn"].submit();

		// TODO: if ap_email exists, do this
		/*document.getElementById('ap_email').value = username;
		document.getElementById('ap_password').value = password;
		document.getElementById('ap_signin_form').submit();*/
	}, loginInfo.username, loginInfo.password);
	this.capture("amazon/images/login-after.png");
}

function getLoginInfo() {
	var file_h = fs.open('amazon/buyer/amazon-login-info.txt', 'r');
	var toReturn = {
		username: file_h.readLine(),
		password: file_h.readLine()
	}
	file_h.close();

	return toReturn
}

function clearCart() {
	this.capture("amazon/images/clearing-cart-before.png");
	try {
		this.clickLabel("Delete");
		this.thenOpen(this.getCurrentUrl(), clearCart);
	} catch (err) {
		this.echo("The cart is empty now");
		this.capture("amazon/images/clearing-cart-after.png");
	}
}

function addProductToCart(a) {
	this.echo("Adding something to the cart." + a.url);
	this.click("#add-to-cart-button");
}

function proceedToCheckout() {
	this.capture("amazon/images/proceed-to-checkout-before.png");
	this.clickLabel("Proceed to checkout");
}

function selectShippingAddress() {
	this.capture("amazon/images/select-shipping-address-before.png");
	// TODO: Needs to be more generic..
	this.click(".ship-to-this-address > span:nth-child(1) > a:nth-child(1)");
}

function selectShippingMethod() {
	this.capture("amazon/images/select-shipping-method-before.png");
	// TODO: Needs to be more generic..
	this.click("div.save-sosp-button-box:nth-child(2) > div:nth-child(1) > span:nth-child(1) > span:nth-child(1) > input:nth-child(1)");
}

function selectPaymentMethod() {
	this.capture("amazon/images/select-payment-method-before.png");
	//this.click("input[value=\"EUR\"]");
	this.click("#continue-top");
}

function placeOrder() {
	this.capture("amazon/images/place-order-before.png");
	//this.click(".place-your-order-button");
}

function logBoughtProducts() {
	this.echo("Logging all products bought");

	var toLog = "";
	for (var i = 0; i < productUrls.length; i++) {
		toLog += productUrls[i] + "\n";
	}
	fs.write("bought-products.txt", toLog, 'a');
}

////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////

var loginInfo = getLoginInfo();

casper.start(amazonLoginPage, login);
casper.thenOpen(amazonCartUrl, clearCart);
for (var i = 0; i < productUrls.length; i++) {
    casper.thenOpen(productUrls[i], addProductToCart);
}

casper.then(proceedToCheckout);
casper.then(selectShippingAddress);
casper.then(selectShippingMethod);
casper.then(selectPaymentMethod);
casper.waitForSelector(".place-your-order-button");
casper.then(placeOrder);
casper.then(logBoughtProducts);

casper.run();
