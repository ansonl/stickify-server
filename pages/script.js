var lastValues = {};

function disableInputs() {
	var progressCode = '<div class="progress"><div class="progress-bar progress-bar-striped active" role="progressbar" aria-valuenow="45" aria-valuemin="0" aria-valuemax="100" style="width: 45%"><span class="sr-only">45% Complete</span></div></div>';
	$('#loadingDiv').html(progressCode)
	$('#formDiv').fadeTo(500, 0.5);
	$('input').prop('disabled', true);
}

function enableInputs() {
	$('#formDiv').fadeTo(500, 1);
	$('input').prop('disabled', false);
	$('#loadingDiv').html('')
}

function submitInput(event) {
	disableInputs();

	var request = $.ajax({
	  method: "POST",
	  url: "http://mich302csd17u:8080/getUser",
	  data: { user: $('#nickname').val(), passcode: $('#pin').val()}
	})
	request.done(function( msg ) {
		if (msg.substring(0,1) == "1") {

			var errorText = '<div class="errorSpan centerHorizSpan">Server provided error: <blockquote>' + msg.substring(2,msg.length) + '</blockquote></div>'

			$('#results').html(errorText);
			enableInputs();
			return;
		}

		var parsed;
		var outputList = '<div>'
		try {
			parsed = JSON.parse(msg)
		} catch (e) {
			var errorText = '<div class="errorSpan centerHorizSpan">Server provided error: <blockquote>' + msg.substring(2,msg.length) + '</blockquote></div>'

			outputList += errorText;

			$('#results').html(errorText);
		}

    for (var i = 0; i < parsed.length; i++) {
    	//golang iotuil.ReadFile issue with % characters
    	//if (i % 3 == 0) {
    	if (i - Math.floor(i / 3) * 3 == 0) {
    		if (i > 0) {
    			outputList += '<div id="sideDivRight" class="col-lg-3"></div></div>';
    		}
    		outputList += '<div class="container"><div id="sideDivLeft" class="col-lg-3"></div>';
    	}

    	var colSize = 6;

    	if (i + 2 < parsed.length) {
    		colSize = 2;
    	} else {
    		//golang iotuil.ReadFile issue with % characters
    		//colSize = 6 / (parsed.length % 3);
    		colSize = 6 / (parsed.length - Math.floor(parsed.length / 3) * 3);
    	}

    	outputList += '<div id="noteDiv" class="col-lg-'  + colSize + '">' + '<ul>'

    	for (var j = 0; j < parsed[i].length; j++) {
    		if (parsed[i][j].length > 0)
    			outputList += '<li>' + parsed[i][j] + '</li>'
    	}

    	outputList += '</ul></div>'

    }

    outputList += '</div>';

    $('#formDiv').fadeOut(500, function () {
    	$('#results').html(outputList);
    });

  });
	request.fail(function( jqXHR, textStatus ) {
	  $('#results').html('Request failed with reason ' + textStatus);
	  enableInputs();
	});

	return false
}
var something = submitInput
$(document).ready(function() {
	
	$('#results').html('');

	$('#userInfoForm').submit(something);
});

