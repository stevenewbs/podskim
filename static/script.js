
function init() {
	console.log("HI");
	$("form").submit(function( event ) {
		addurl($(this));	
		event.preventDefault();
	});
	//alert("hi");
}

function addurl(obj) {
	var url = obj.children("input[name='newurl']").val();
	console.log(url);
	if (url.trim() == "") {
		ShowAlert("Please enter a full URL");
		console.log("Input is empty");
	}
	else {
		ShowAlert("Adding..." + url);
		console.log("Url is " + url);	
	}
}
var tovar;

function ShowAlert(string) {
	$('.alert').fadeIn('fast').children('p').text(string);
	tovar = setTimeout(CloseAlert, 3500);
}

function CloseAlert() {
	clearTimeout(tovar);
	$('.alert').fadeOut('fast');
}
