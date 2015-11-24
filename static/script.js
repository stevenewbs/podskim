
function init() {
	window.onbeforeunload = function (e) {
		//alert("Hi");
		Quit();
  	return "bibble"
	};

	$("form").submit(function( event ) {
		addurl($(this));
		event.preventDefault();
	});
	//alert("hi");
	$('.feed').each(function() {
		var id = $(this).attr('id')
		var name = $(this).attr('data-name')
		getfeed(id, name, 1)
	});
}

function reload() {
	window.location.reload()
}

function addurl(obj) {
	var url = obj.children("input[name='newurl']").val();
	var name = obj.children("input[name='name']").val();
	console.log(url);
	if (url.trim() == "") {
		ShowAlert("Please enter a full URL");
		console.log("Input is empty");
	}
	else {
		//ShowAlert("Adding..." + url);
		//console.log("Url is " + url);
		$.post("/add", {newurl: url, name: name }, "json").done(function (json) {
                        //alert("done");
                        reload()
                    }).fail(function (jqxhr, textStatus, error) {
                        var error = textStatus + ": " + error;
                        alert(error);
                    });
	}
}

function removeurl(name) {
	console.log(name);
	y = confirm("Are you sure you want to remove "+ name + "?")
	if (y == true) {
		//ShowAlert("Adding..." + url);
		console.log("Removing " + name);
		$.post("/delete", {name: name }, "json").done(function (json) {
			//alert("done");
			reload()
		}).fail(function (jqxhr, textStatus, error) {
			var error = textStatus + ": " + error;
			alert(error);
		});
	}
}

function getfeed(id, name, amount) {
	console.log("Getting "+ name)

	$('#' + id).load("/feed .feeddata", {name: name, amount: amount}, function () {
		//alert("done");
	});
}

function ShowAlert(string) {
	alert(string);
}

function Quit() {
	$.get("/quit");
}
