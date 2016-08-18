var g_menuShow = true;
var GKey = {
	init: function() {
		var pass = getCookie('psw');
		var pubKey = localStorage.getItem('PubKey');
		if (pubKey)
			this.Public = pubKey;
		if (pass && localStorage.getItem('EncKey')) {
			this.decrypt(localStorage.getItem('EncKey'), pass)
		}
	}, 
	decrypt: function( encKey, pass ) {
		var decrypted = CryptoJS.AES.decrypt(encKey, pass).toString(CryptoJS.enc.Hex);
		var prvkey = '';
		for ( i=0; i < decrypted.length; i+=2 ) {
			var num = parseInt( decrypted.substr(i,2),16);
			prvkey += String.fromCharCode(num);
		}
		if (this.verify(prvkey, this.Public)) {
			this.Private = prvkey;
			this.Password = pass;
			return true;
		}
		return false;
	},
	save: function() {
		var encryptedAES = CryptoJS.AES.encrypt(this.Private, this.Password);
		localStorage.setItem('EncKey', encryptedAES );
		localStorage.setItem('PubKey', GKey.Public );
		setCookie('psw', this.Password);
	},
	verify: function( prvkey, pubkey ) {
  		var sigalg = 'SHA256withECDSA';
		var msg = 'test';
  		var sig = new KJUR.crypto.Signature({"alg": sigalg});

  		sig.initSign({'ecprvhex': prvkey, 'eccurvename': this.Curve});
  		sig.updateString(msg);
  		var sigval = sig.sign();

  		var siga = new KJUR.crypto.Signature({"alg": sigalg, "prov": "cryptojs/jsrsa"});
  		siga.initVerifyByPublicKey({'ecpubhex': pubkey, 'eccurvename': this.Curve});
  		siga.updateString(msg);
  		return siga.verify(sigval);
	},
	Curve: 'secp256r1',
	Password: '',
	Private: '',
	Public:  '',
}

GKey.init();

function getCookie(name) {
	var matches = document.cookie.match(new RegExp(
    	"(?:^|; )" + name.replace(/([\.$?*|{}\(\)\[\]\\\/\+^])/g, '\\$1') + "=([^;]*)"
  	));
  	return matches ? decodeURIComponent(matches[1]) : undefined;
}

function setCookie(name, value, options) {
	options = options || {};
	var expires = options.expires;

 	if (typeof expires == "number" && expires) {
    	var d = new Date();
    	d.setTime(d.getTime() + expires * 1000);
    	expires = options.expires = d;
  	}
	if (expires && expires.toUTCString) {
    	options.expires = expires.toUTCString();
  	}
	value = encodeURIComponent(value);
	var updatedCookie = name + "=" + value;

  	for (var propName in options) {
    	updatedCookie += "; " + propName;
    	var propValue = options[propName];
    	if (propValue !== true) {
      		updatedCookie += "=" + propValue;
    	}
  	}
	document.cookie = updatedCookie;
}

function Demo() {
	var id = $("#demo");
	var val = id.val();
	if (val == 0) {
		id.prev().find("em").html("Hide opportunities");
		id.val(1);
		$("body").addClass("demoMode");
	} else {
		id.prev().find("em").html("Show opportunities");
		id.val(0);
		$("body").removeClass("demoMode");
	}
}

var obj;

function Alert(title, text, type) {
	obj.css({"position":"relative"});
	var id = obj.parents(".modal").attr("id");
	swal({
		title : title,
		text : text,
		allowEscapeKey : false,
		type : type
	}, function (isConfirm) {
		if (isConfirm) {
			$("#" + id).modal("hide");
			obj.removeClass("whirl standard");
		}
	});
	$(".sweet-alert").appendTo(obj);
}

function preloader(elem) {
	obj = $("#" + elem.id).parents("[data-sweet-alert]");
	if (!obj.find(".sk-cube-grid").length) {
		obj.append('<div class="sk-cube-grid"><div class="sk-cube sk-cube1"></div><div class="sk-cube sk-cube2"></div><div class="sk-cube sk-cube3"></div><div class="sk-cube sk-cube4"></div><div class="sk-cube sk-cube5"></div><div class="sk-cube sk-cube6"></div><div class="sk-cube sk-cube7"></div><div class="sk-cube sk-cube8"></div><div class="sk-cube sk-cube9"></div></div>');
	}
}

function dl_navigate (page, parameters) {
    var json = JSON.stringify(parameters);
    //$('#loader').spin();
    $.post("content?controllerHTML="+page, { tpl_name: page, parameters: json },
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
	if (g_menuShow) {
	    parametersJson = "";
	    if (typeof lang!='undefined') {
	        parametersJson: '{"lang":"1"}'
	    }
	    $("#dl_menu").load( "content?controllerHTML=menu", { parameters: parametersJson }, function() {
	    });
	} else {
		$("#dl_menu").html('');
	}
}

function login_ok (result) {

    if (result=='1') {

        console.log('login_ok=1');

        $('#myModal').modal('hide');
        $('#myModalLogin').modal('hide');
        $('.modal-backdrop').remove();
        $('.modal-backdrop').css('display', 'none');

        if (typeof(get_key_and_sign)==='undefined' || get_key_and_sign=='null') {

            var tpl_name = $('#tpl_name').val();
            if (!tpl_name || typeof(tpl_name)==='undefined' || tpl_name=='installStep0' || tpl_name=='installStep6')
                tpl_name = 'home';

            console.log('tpl_name = ', tpl_name);

            if ($("#mobileos").val() == "1") {
                $("#page-wrapper").css('padding-bottom', '40px');
                $(".navbar-default").css('border-color', '#ccc');
                $("#ios_menu").css('display', 'block');
            }

            $( "#dl_content" ).load( "content", { tpl_name: tpl_name}, function() {
					$("#main-login").html('');
					$("#loader").spin(false);
            });
        }
        else if (get_key_and_sign=='sign') {
            console.log('get_key_and_sign=sign');
            doSign('sign');
            $("#main-login").html('');
            $("#loader").spin(false);
        }
        else if (get_key_and_sign=='send_to_net') {
            console.log('get_key_and_sign=send_to_net');
            doSign('sign');
            $("#send_to_net").trigger("click");
            $("#main-login").html('');
            $("#loader").spin(false);
        }
		g_menuShow = true;
		load_menu();
    }
    else if (result=='not_available') {
        $("#modal_alert").html('<div id="alertModalPull" class="alert alert-danger alert-dismissable"><button type="button" class="close" data-dismiss="alert" aria-hidden="true">×</button><p>'+$('#pool_is_full').val()+'</p></div>');
        $("#loader").spin(false);
    }
    else {
        $("#modal_alert").html('<div id="alertModalPull" class="alert alert-danger alert-dismissable"><button type="button" class="close" data-dismiss="alert" aria-hidden="true">×</button><p>'+$('#incorrect_key_or_password').val()+'</p></div>');
        $("#loader").spin(false);
    }
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

function save_key () {

    $('#loader').spin();
    console.log("$('#loader').spin();");
    $('#modal_alert').html( "" );
    $('#key').text( $("#modal_key").val() );
    $('#password').text( $("#modal_password").val() );
    $('#setup_password').text( $("#modal_setup_password").val() );
    if ($("#modal_save_key").is(':checked')) {
        console.log("save_key 1")
        $('#save_key').text("1");
    } else {
        console.log("save_key 0");
        console.log("modal_save_key:", $("#modal_save_key").val())
        $('#save_key').text("0");
    }
}

function base_convert(number, frombase, tobase) {
    return parseInt(number + '', frombase | 0)
        .toString(tobase | 0);
}

function img2key(img, key_id) {

    //console.log(img);
    var image = new Image();
    image.src = img;
    image.onload = function() {

        $('#canvas_key').attr('width', this.width);
        $('#canvas_key').attr('height', this.height);
        var c=document.getElementById("canvas_key");
        var ctx=c.getContext("2d");

        ctx.drawImage(image,0,0);

        // вначале прочитаем инфу, где искать rsa-ключ (64 пиксла = 64 бита = 8 байт = 4 числа = x,y,w,h)
        var count_bits = 0;
        var byte = '';
        var rsa_search_params = [];
        for (var x=0; x<64; x++) {
            var Pixel = ctx.getImageData(x, 0, 1, 1);
            //console.log(x+' '+y+' / '+Pixel.data[0]+' '+Pixel.data[1]+' '+Pixel.data[2]);
            if (Pixel.data[0] > 100)
                var bin = 1;
            else
                var bin = 0;
            byte = byte+''+bin;
            count_bits=count_bits+1;
            if (count_bits==16) {
                //console.log(byte+' == '+base_convert(byte, 2, 10));
                rsa_search_params.push(base_convert(byte, 2, 10));
                count_bits = 0;
                byte = '';
            }
        }
        console.log(rsa_search_params);

        var hex = '';
        var count_bits = 0;
        var byte = '';
        var hex_byte = '';
        for (var y=rsa_search_params[1]; y<(Number(rsa_search_params[1])+Number(rsa_search_params[3])); y++) {
            for (var x=rsa_search_params[0]; x<(Number(rsa_search_params[0])+Number(rsa_search_params[2])-1); x++) {
                var Pixel = ctx.getImageData(x, y, 1, 1);
                //console.log(x+' '+y+' / '+Pixel.data[0]+' '+Pixel.data[1]+' '+Pixel.data[2]);
                if (Pixel.data[0] > 100)
                    var  bin = 1;
                else
                    var bin = 0;
                byte = byte+''+bin;
                count_bits=count_bits+1;
                if (count_bits==8) {
                    hex_byte = strpadleft(base_convert(byte, 2, 16));
                    //console.log(byte+'='+hex_byte);
                    hex = hex + ''+ hex_byte;
                    count_bits = 0;
                    byte = '';
                }
            }
        }
        hex = hex.split('00000000');
        console.log(hex);
        var key = hexToBase64(hex[0]);
        console.log(key);
        $('#'+key_id).val(key);
    };
}

function strpadleft(mystr) {
   // mystr = dechex(mystr);
    var pad = "00";
    var str = "" + mystr;
    return (pad.substring(0, pad.length - str.length) + str);
}

function hexToBase64(str) {
    return btoa(String.fromCharCode.apply(null,
            str.replace(/\r|\n/g, "").replace(/([\da-fA-F]{2}) ?/g, "0x$1 ").replace(/ +$/, "").split(" "))
    );
}

function base64ToHex(str) {
    for (var i = 0, bin = atob(str.replace(/[ \r\n]+$/, "")), hex = []; i < bin.length; ++i) {
        var tmp = bin.charCodeAt(i).toString(16);
        if (tmp.length === 1) tmp = "0" + tmp;
        hex[hex.length] = tmp;
    }
    return hex.join(" ");
}


function hex2a(hex) {
    var str = '';
    for (var i = 0; i < hex.length; i += 2)
        str += String.fromCharCode(parseInt(hex.substr(i, 2), 16));
    return str;
}

function get_e_n_sign(key, pass, forsignature, alert_div) {

    var modulus = '';
    var exp = '';
    var hSig = '';
    var decrypt_PEM = '';
    key = key.trim();
    // ключ может быть незашифрованным, но без BEGIN RSA PRIVATE KEY
    if (key.substr(0,4) == 'MIIE')
        decrypt_PEM = '-----BEGIN RSA PRIVATE KEY-----'+key+'-----END RSA PRIVATE KEY-----';
    else if (pass && key.indexOf('RSA PRIVATE KEY')==-1) {
        try{

            ivAndText = atob(key);
            iv = ivAndText.substr(0, 16);
            encText = ivAndText.substr(16);
            cipherParams = CryptoJS.lib.CipherParams.create({
                ciphertext: CryptoJS.enc.Base64.parse(btoa(encText))
            });
            pass = CryptoJS.enc.Latin1.parse(hex_md5(pass));
            var decrypted = CryptoJS.AES.decrypt(cipherParams, pass, {mode: CryptoJS.mode.CBC, iv: CryptoJS.enc.Utf8.parse(iv), padding: CryptoJS.pad.Iso10126 });
            decrypt_PEM = hex2a(decrypted.toString());

/*
            cipherParams = CryptoJS.lib.CipherParams.create({
                ciphertext: CryptoJS.enc.Base64.parse((key.replace(/\n|\r/g, "")))
            });
            key = CryptoJS.enc.Latin1.parse(hex_md5(pass))
            var decrypted = CryptoJS.AES.decrypt(cipherParams, key, {mode: CryptoJS.mode.CBC, iv: CryptoJS.enc.Base64.parse("AAAAAAAAAAAAAAAAAAAAAA=="), padding: CryptoJS.pad.NoPadding });
            var decrypt_PEM = hex2a(decrypted.toString());
*/


        } catch(e) {
            console.log(e)
           decrypt_PEM = 'invalid base64 code';
       }
    }
    else
        decrypt_PEM = key;
    console.log('decrypt_PEM='+decrypt_PEM);
    console.log('typeof decrypt_PEM ='+typeof decrypt_PEM );
   if (typeof decrypt_PEM != "string" || decrypt_PEM.indexOf('RSA PRIVATE KEY')==-1) {
       console.log('incorrect_key_or_password');
        $("#loader").spin(false);
        $("#"+alert_div).html('<div id="alertModalPull" class="alert alert-danger alert-dismissable"><button type="button" class="close" data-dismiss="alert" aria-hidden="true">×</button><p>'+$('#incorrect_key_or_password').val()+'</p></div>');
       console.log(alert_div);
    }
    else {
        var rsa = new RSAKey();
        rsa.readPrivateKeyFromPEMString(decrypt_PEM);
        var a = rsa.readPrivateKeyFromPEMString(decrypt_PEM);
        modulus = a[1];
        exp = a[2];

        if (forsignature!='') {
            console.log('forsignature='+forsignature);
            hSig = rsa.signString(forsignature, 'sha1');
            console.log('hSig='+hSig);
        }

        delete rsa;
    }
    var data = new Object();
    data['modulus'] = modulus;
    data['exp'] = exp;
    data['hSig'] = hSig;
    data['decrypt_key'] = decrypt_PEM;
    return data;
}