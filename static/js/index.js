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
    $( "#dl_menu" ).load( "content?controllerHTML=menu", { parameters: parametersJson }, function() {
    });
}

function doSign_(type) {

    if (typeof(type) === 'undefined') type = 'sign';

    console.log('type=' + type);

    var SIGN_LOGIN = false;

    jQuery.extend({
        getValues: function (url) {
            var result = null;
            $.ajax({
                url: url,
                type: 'get',
                dataType: 'json',
                async: false,
                success: function (data) {
                    result = data;
                }
            });
            return result;
        }
    });

    var key = $("#key").text();
    var pass = $("#password").text();
    var setup_password = $("#setup_password").text();
    var save_key = $("#save_key").text();
    console.log("save_key=" + save_key);

    if (key.length < 512) {
        $("#modal_alert").html('<div id="alertModalPull" class="alert alert-danger alert-dismissable"><button type="button" class="close" data-dismiss="alert" aria-hidden="true">×</button><p>'+$('#incorrect_key_or_password').val()+'</p></div>');
        $("#loader").spin(false);
        return false;
    }
    if (type=='sign') {
        var forsignature = $("#for-signature").val();
    }
    else {
        if (key) {
            // авторизация с ключем и паролем
            if ($('#exchangeTemplate').val() == "1") {
                var forsignature = $.getValues("ajax?controllerName=ESignLogin");
            } else {
                var forsignature = $.getValues("ajax?controllerName=signLogin");
            }
            SIGN_LOGIN = true;
        }
    }

    console.log('forsignature='+forsignature);

    if (forsignature) {
        console.log('key='+key);
        console.log('pass='+pass);
		if (key != localStorage.getItem('dcoin_key')) {
			localStorage.setItem('dcoin_pass', pass );
			localStorage.setItem('dcoin_key', key );
		}
        var e_n_sign = get_e_n_sign(key, pass, forsignature, 'modal_alert');
	} else {
		return;
	}
	if (SIGN_LOGIN) {

			console.log('SIGN_LOGIN');

			//$("#loader").spin();
			if (key) {
                var privKey = "";
                if (save_key == "1") {
                    privKey = e_n_sign['decrypt_key']
                }

                if ($('#exchangeTemplate').val() == "1") {
                    var check_url = 'ajax?controllerName=ECheckSign'
                } else {
                    var check_url = 'ajax?controllerName=check_sign'
                }
				// шлем подпись на сервер на проверку
				$.post( check_url, {
							'sign': e_n_sign['hSig'],
							'n' : e_n_sign['modulus'],
							'e': e_n_sign['exp'],
                            'private_key': privKey,
                            'forsignature' : forsignature,
                            'setup_password': setup_password
						}, function (data) {
							// залогинились
							console.log("data.result: ", data.result);
							login_ok( data.result );

						}, 'JSON'
				);
			}
			else {

				hash_pass = hex_sha256(hex_sha256(pass));
				// шлем хэш пароля на проверку и получаем приватный ключ
				$.post( 'ajax?controllerName=check_pass', {
							'hash_pass': hash_pass
						}, function (data) {
							// залогинились
							login_ok( data.result );

							$("#modal_key").val(data.key);
							$("#key").text(data.key);
							//alert(data.key);

						}, 'JSON'
				);

			}

			//$("#loader").spin(false);

	}
	else {
			$("#signature1").val(e_n_sign['hSig']);
	}
}