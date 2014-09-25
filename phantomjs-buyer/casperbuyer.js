var stepNum = 0;
var captureAllSteps = true;
	
var fs = require('fs');
var utils = require("utils");
var casper = require('casper').create({
	onStepComplete: function() {
		this.echo("Finished step " + stepNum + ". Title: " + this.getTitle());
		if (captureAllSteps) {
			this.capture("images/" + stepNum + '.png');
		}
		stepNum++;
	}
});
casper.userAgent('Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)');
//phantom.cookiesEnabled = true;

////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////

var amazonLoginPage = 'https://www.amazon.com/ap/signin?_encoding=UTF8&openid.assoc_handle=usflex&openid.claimed_id=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0%2Fidentifier_select&openid.identity=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0%2Fidentifier_select&openid.mode=checkid_setup&openid.ns=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0&openid.ns.pape=http%3A%2F%2Fspecs.openid.net%2Fextensions%2Fpape%2F1.0&openid.pape.max_auth_age=0&openid.return_to=https%3A%2F%2Fwww.amazon.com%2Fgp%2Fyourstore%2Fhome%3Fie%3DUTF8%26ref_%3Dnav_custrec_signin';
//var amazonLoginPage = 'http://casperjs.org';
var productUrls =  casper.cli.args;

////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////

function login(username, password) {
	document.getElementById('ap_email').value = username;
	document.getElementById('ap_password').value = password;
	document.getElementById('ap_signin_form').submit();
}

function getLoginInfo() {
	var file_h = fs.open('amazon-login-info.txt', 'r');
	var toReturn = {
		username: file_h.readLine(),
		password: file_h.readLine()
	}
	file_h.close();
	
	return toReturn
}

function addProductToCart(a) {
	this.echo("Adding something to the cart." + a.url);
	this.click("#add-to-cart-button");
}

function proceedToCheckout() {
	this.capture("images/a-before-proceedtocheckout.png");
	this.clickLabel("Proceed to checkout");
}

function selectShippingAddress() {
	this.capture("images/b-before-selectshippingaddress.png");
	// TODO: Needs to be more generic..
	this.click(".ship-to-this-address > span:nth-child(1) > a:nth-child(1)");
}

function selectShippingMethod() {
	this.capture("images/c-before-selectshippingmethod.png");
	// TODO: Needs to be more generic..
	this.click("div.save-sosp-button-box:nth-child(2) > div:nth-child(1) > span:nth-child(1) > span:nth-child(1) > input:nth-child(1)");
}

function selectPaymentMethod() {
	this.capture("images/d-before-selectpayment.png");
	//this.click("input[value=\"EUR\"]");
	this.click("#continue-top");
}

function placeOrder() {
	this.capture("images/e-before-placeorder.png");
	this.click(".place-your-order-button");
}

function logBoughtProducts() {
	this.echo("Logging all products bought");
	// TODO: Do this... :)
}

////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////

var loginInfo = getLoginInfo();

casper.start(amazonLoginPage);
casper.thenEvaluate(login, loginInfo.username, loginInfo.password);
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
//casper.exit();
