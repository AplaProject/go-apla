function dl_navigate (page, parameters) {
    var json = JSON.stringify(parameters);
    //$('#loader').spin();
    $.post("content?controllerName="+page, { tpl_name: page, parameters: json },
        function(data) {
            //$("#loader").spin(false);
            $('#dl_content').html( data );
			/*if ( parameters && parameters.hasOwnProperty("lang")) {
				if ( page[0] == 'E' )
					load_emenu();
				else
					load_menu();
			}*/
            window.scrollTo(0,0);
        }, "html");
}

function load_menu(lang) {
    parametersJson = "";
    if (typeof lang!='undefined') {
        parametersJson: '{"lang":"1"}'
    }
    $( "#dl_menu" ).load( "ajax?controllerName=menu", { parameters: parametersJson }, function() {
    });
}