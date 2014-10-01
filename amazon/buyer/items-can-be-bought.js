var stepNum = 0;
var captureAllSteps = false;

var fs = require('fs');
var utils = require("utils");
var casper = require('casper').create({
	onStepComplete: function() {
		if (captureAllSteps) {
			this.capture("amazon/images/" + stepNum + '.png');
		}
		stepNum++;
	}
});
casper.userAgent('Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)');

////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////

var amazonLoginPage = 'https://www.amazon.com/ap/signin?_encoding=UTF8&openid.assoc_handle=usflex&openid.claimed_id=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0%2Fidentifier_select&openid.identity=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0%2Fidentifier_select&openid.mode=checkid_setup&openid.ns=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0&openid.ns.pape=http%3A%2F%2Fspecs.openid.net%2Fextensions%2Fpape%2F1.0&openid.pape.max_auth_age=0&openid.return_to=https%3A%2F%2Fwww.amazon.com%2Fgp%2Fyourstore%2Fhome%3Fie%3DUTF8%26ref_%3Dnav_custrec_signin';
var productUrls =  casper.cli.args;

////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////

function login(username, password) {
    // TODO: if auth-email exists, do this
    document.getElementById('auth-email').value = username;
    document.getElementById('auth-password').value = password;
    document.forms["signIn"].submit();

    // TODO: if ap_email exists, do this
    /*document.getElementById('ap_email').value = username;
    document.getElementById('ap_password').value = password;
    document.getElementById('ap_signin_form').submit();*/
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

function doesShip() {
	// This item ships to Gothenburg, Sweden.
	// This item does not ship to Gothenburg, Sweden.

	var re = new RegExp("This item ships to");
	html = casper.getHTML();
	return re.test(html);
}

function needsMoreInput() {
	html = casper.getHTML();
	var re = new RegExp("To buy, select");
	return re.test(html);
}

function seeIfProductCanBeBought() {
	var ships = doesShip();
	var doesNeedMoreInput = needsMoreInput();

	var response;
	if (!ships || doesNeedMoreInput) {
		response = 1;
	} else {
		response = 0;
	}

	this.echo(this.getCurrentUrl() + ";" + response);
}

////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////

var loginInfo = getLoginInfo();

casper.start(amazonLoginPage);
casper.thenEvaluate(login, loginInfo.username, loginInfo.password);

for (var i = 0; i < productUrls.length; i++) {
	casper.thenOpen(productUrls[i]);
	casper.then(seeIfProductCanBeBought);
}

casper.run();
