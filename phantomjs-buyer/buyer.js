var fs = require('fs');
var system = require('system');
var args = system.args;

if (args.length !== 2) {
  console.log('You must specify a product url');
  phantom.exit()
}

var page = new WebPage(), testindex = 0, loadInProgress = false;
var productUrl = args[1];




page.onConsoleMessage = function(msg) {
  console.log("#### " + msg);
};
page.onLoadStarted = function() {
	loadInProgress = true;
};

page.onLoadFinished = function() {
	loadInProgress = false;
};

buy(productUrl);

function buy(url) {
	visit(url, findAndClickSignIn);
}

function visit(url, onLoad) {
	console.log("Visiting " + url);
	page.settings.userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_2) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.6 Safari/537.11";
	page.open(url, function(status) {
		console.log(status);
		
		var interval = setInterval(function(){
			if(!loadInProgress) {
				clearInterval(interval);
				
				onLoad();
			}
		},50);
		
	});
}

function findAndClickSignIn() {
	page.render("images/product-page.png");
	var signInLink = page.evaluate(function() {
	  return document.querySelector(".oneClickSignInLink");
	});
	console.log("Signing in to " + signInLink.href);		
	
	visit(signInLink.href, signIn);
}

function signIn() {
	page.render("images/sign-in-page.png");
	
	loginInfo = getLoginInfo();
	page.evaluate(function(username, password) {
		document.getElementById('ap_email').value = username;
		document.getElementById('ap_password').value = password;
		document.getElementById('ap_signin_form').submit();
	}, loginInfo.username, loginInfo.password);
	page.render("images/sign-in-page--data-entered.png");
	
	var interval = setInterval(function(){
		if(!loadInProgress) {
		   clearInterval(interval);
		   
		   page.render("images/submitted.png");
		   findAndClickOneClickBuy();
		}
	},50);
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

function findAndClickOneClickBuy() {
	page.evaluate(function() {
		document.getElementById("one-click-button").click();
		//document.getElementById("one-click-button").outerHTML = "Tjabba!";
	});
	
	var interval = setInterval(function(){
		if(!loadInProgress) {
		   clearInterval(interval);
		   
		   page.render("images/one-click-pressed.png");
		   phantom.exit();
		}
	},50);
}




/**
* Function : dump()
* Arguments: The data - array,hash(associative array),object
*    The level - OPTIONAL
* Returns  : The textual representation of the array.
* This function was inspired by the print_r function of PHP.
* This will accept some data as the argument and return a
* text that will be a more readable version of the
* array/hash/object that is given.
*/
function dump(arr,level) {
var dumped_text = "";
if(!level) level = 0;

//The padding given at the beginning of the line.
var level_padding = "";
for(var j=0;j<level+1;j++) level_padding += "    ";

if(typeof(arr) == 'object') { //Array/Hashes/Objects
 for(var item in arr) {
  var value = arr[item];
 
  if(typeof(value) == 'object') { //If it is an array,
   dumped_text += level_padding + "'" + item + "' ...\n";
   dumped_text += dump(value,level+1);
  } else {
   dumped_text += level_padding + "'" + item + "' => \"" + value + "\"\n";
  }
 }
} else { //Stings/Chars/Numbers etc.
 dumped_text = "===>"+arr+"<===("+typeof(arr)+")";
}
return dumped_text;
} 
