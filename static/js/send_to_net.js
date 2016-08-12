function send_to_net_success(data){
	if (typeof data.error != "undefined") {
		Alert("Error", data.error, "error");
	} else {
		interval = setInterval(function() {
			$.ajax({
				type: 'POST',
				url: 'ajax?controllerName=txStatus',
				data: {
					'hash': data.hash
				},
				dataType: 'json',
				crossDomain: true,
				success: function(txStatus){

					console.log("txStatus", txStatus);

					if (typeof txStatus.wait != "undefined") {
						console.log("txStatus", txStatus);
					} else if (typeof txStatus.error != "undefined") {
						Alert("Error", txStatus.error, "error");
						clearInterval(interval);
					} else {
						clearInterval(interval);
						Alert("Success", "", "success");
					}
				},
				error: function(xhr, status, error) {
					clearInterval(interval);
					Alert("Error Money transfer", error, "error");
				},
			});
		}, 1000)
	}
}