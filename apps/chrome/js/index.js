
function dc_navigate (page, parameters) {

    var json = JSON.stringify(parameters);

    $('#loader').spin();
    $.post("content?page="+page, { tpl_name: page, parameters: json },
        function(data) {
            $("#loader").spin(false);
            //console.log('$("#loader").spin(false)');
            $('#dc_content').html( data );
			if ( parameters && parameters.hasOwnProperty("lang")) {
				if ( page[0] == 'E' )
					load_emenu();
				else
					load_menu();
			}
            window.scrollTo(0,0);
            if ($(".sidebar-collapse").is(":visible") && $(".navbar-toggle").is(":visible")) {
               $('.sidebar-collapse').collapse('toggle');
            }
        }, "html");

}

function map_navigate (page) {

    $('#loader').spin();
    $.getScript("https://maps.googleapis.com/maps/api/js?sensor=false", function(){

        $( "#dc_content" ).load( "content", { tpl_name: page }, function() {
            $('#loader').spin(false);
            window.scrollTo(0,0);
            if ($(".sidebar-collapse").is(":visible") && $(".navbar-toggle").is(":visible")) {
                $('.sidebar-collapse').collapse('toggle');
            }
        });
    });

}

function user_photo_navigate (page) {

    $('#loader').spin();
    $.getScript("static/js/jquery.webcam.as3.js", function(){

                $( "#dc_content" ).load( "content", { tpl_name: page }, function() {
                    $.getScript("static/js/sAS3Cam.js", function(){
                        $('#loader').spin(false);
                    });
                });

    });

    window.scrollTo(0,0);
    if ($(".sidebar-collapse").is(":visible") && $(".navbar-toggle").is(":visible")) {
        $('.sidebar-collapse').collapse('toggle');
    }

}

function user_webcam_navigate (page) {

    $('#loader').spin();
        $( "#dc_content" ).load( "content", { tpl_name: page }, function() { });
        $('#loader').spin(false);
        window.scrollTo(0,0);
        if ($(".sidebar-collapse").is(":visible") && $(".navbar-toggle").is(":visible")) {
            $('.sidebar-collapse').collapse('toggle');
        }
}

function load_menu(lang) {
    parametersJson = "";
    if (typeof lang!='undefined') {
        parametersJson: '{"lang":"1"}'
    }
    $( "#dc_menu" ).load( "ajax?controllerName=menu", { parameters: parametersJson }, function() {
       // $( "#dc_content" ).load( "content", { }, function() {
            $.getScript("static/js/plugins/metisMenu/metisMenu.min.js", function(){
                $.getScript("static/js/sb-admin.js");
            });
        //});
    });
}

function load_emenu(lang) {
    parametersJson = "";
    if (typeof lang!='undefined') {
        parametersJson: '{"lang":"1"}'
    }
    $( "#dc_emenu" ).load( "ajax?controllerName=EMenu", { parameters: parametersJson }, function() {
            $.getScript("static/js/plugins/metisMenu/metisMenu.min.js", function(){
                $.getScript("static/js/sb-admin.js");
            });
    });
}


function login_ok (result) {

    if (result=='1') {

        console.log('login_ok=1');

        $('#myModal').modal('hide');
        $('#myModalLogin').modal('hide');
        $('.modal-backdrop').remove();
        $('.modal-backdrop').css('display', 'none');

        if ($('#exchangeTemplate').val() == "1") {

            $('#SigninKeyModal').modal('hide');
            window.location.href = $('#EHost').val();

        } else if (typeof(get_key_and_sign)==='undefined' || get_key_and_sign=='null') {

            var tpl_name = $('#tpl_name').val();
            if (!tpl_name || typeof(tpl_name)==='undefined' || tpl_name=='installStep0' || tpl_name=='installStep6')
                tpl_name = 'home';

            if ($("#mobileos").val() == "1") {
                $("#page-wrapper").css('padding-bottom', '40px');
                $(".navbar-default").css('border-color', '#ccc');
                $("#ios_menu").css('display', 'block');
            }
            $( "#dc_menu" ).load( "ajax?controllerName=menu", { }, function() {
                $( "#dc_content" ).load( "content", { tpl_name: tpl_name}, function() {
                    $.getScript("static/js/plugins/metisMenu/metisMenu.min.js", function() {
                        $.getScript("static/js/sb-admin.js");
                        $("#main-login").html('');
                        $("#loader").spin(false);
                    });
                });
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


function logout () {
	localStorage.removeItem('dcoin_pass');
	localStorage.removeItem('dcoin_key');
    $.get("ajax?controllerName=logout",
        function() {
            window.location.href = "/";
        });
}

var keyStr = "ABCDEFGHIJKLMNOP" +
    "QRSTUVWXYZabcdef" +
    "ghijklmnopqrstuv" +
    "wxyz0123456789+/" +
    "=";

function decode64(input) {
    var output = "";
    var chr1, chr2, chr3 = "";
    var enc1, enc2, enc3, enc4 = "";
    var i = 0;

    // remove all characters that are not A-Z, a-z, 0-9, +, /, or =
    var base64test = /[^A-Za-z0-9\+\/\=]/g;
    if (base64test.exec(input)) {
        alert("There were invalid base64 characters in the input text.\n" +
            "Valid base64 characters are A-Z, a-z, 0-9, '+', '/',and '='\n" +
            "Expect errors in decoding.");
    }
    input = input.replace(/[^A-Za-z0-9\+\/\=]/g, "");

    do {
        enc1 = keyStr.indexOf(input.charAt(i++));
        enc2 = keyStr.indexOf(input.charAt(i++));
        enc3 = keyStr.indexOf(input.charAt(i++));
        enc4 = keyStr.indexOf(input.charAt(i++));

        chr1 = (enc1 << 2) | (enc2 >> 4);
        chr2 = ((enc2 & 15) << 4) | (enc3 >> 2);
        chr3 = ((enc3 & 3) << 6) | enc4;

        output = output + String.fromCharCode(chr1);

        if (enc3 != 64) {
            output = output + String.fromCharCode(chr2);
        }
        if (enc4 != 64) {
            output = output + String.fromCharCode(chr3);
        }

        chr1 = chr2 = chr3 = "";
        enc1 = enc2 = enc3 = enc4 = "";

    } while (i < input.length);

    return unescape(output);
}



function map_init (lat, lng, map_canvas, drag, clickmarker) {

    $("#"+map_canvas).css("display", "block");

    var point = new google.maps.LatLng(lat, lng);
    var mapOptions = {
        center: point,
        zoom: 15,
        mapTypeId: google.maps.MapTypeId.ROADMAP,
        streetViewControl: false
    };
    map = new google.maps.Map(document.getElementById(map_canvas), mapOptions);

    var marker = new google.maps.Marker({
        position: point,
        map: map,
        draggable: drag,
        title: 'You'
    });

    google.maps.event.trigger(map, 'resize');

    google.maps.event.addListener(marker, "dragend", function() {
         set_lat_lng();
    });

    if (clickmarker) {
        google.maps.event.addListener(map, 'click', function(event) {
            placeMarker(event.latLng);
        });
    }

    function placeMarker(location) {

        if (marker == undefined){
            marker = new google.maps.Marker({
                position: location,
                map: map,
                animation: google.maps.Animation.DROP,
            });
        }
        else{
            marker.setPosition(location);
        }
        map.setCenter(location);

        set_lat_lng();

    }

    function set_lat_lng() {

        var lat = marker.getPosition().lat();
        lat = lat.toFixed(5);
        var lng = marker.getPosition().lng();
        lng = lng.toFixed(5);
        console.log(lat, lng)
        document.getElementById('latitude').value = lat;
        document.getElementById('longitude').value = lng;

    }

    marker.setMap(map);
}

function check_key_and_show_modal() {
    if ( $('#key').text().length < 256 ) {
        $('#myModal').modal({ backdrop: 'static' });
    }
}

function check_key_and_show_modal2() {
    console.log('check_key_and_show_modal2');
    if ( $('#key').text().length < 256 ) {
        $('#myModal').modal({ backdrop: 'static' });
    }
    else {
        if (typeof(get_key_and_sign)==='undefined' || get_key_and_sign=='null') {

        }
        else if (get_key_and_sign=='sign') {
            doSign('sign');
        }
        else if (get_key_and_sign=='send_to_net') {
            doSign('sign');
            $("#send_to_net").trigger("click");
        }
    }
}



function send_crop (type, coords, img_id) {

    $.post('ajax/crop_photo.php', {'type' : type, 'coords' : $('#'+coords).text() },
        function(data) {

            $('#'+img_id).html('<img width="350" src="'+data.url+'?r='+Math.random()+'" id="'+type+'">');

        }, "json");

}

function strpadleft(mystr) {
   // mystr = dechex(mystr);
    var pad = "00";
    var str = "" + mystr;
    return (pad.substring(0, pad.length - str.length) + str);
}
function dechex(number) {
    if (number < 0) {
        number = 0xFFFFFFFF + number + 1;
    }
    return parseInt(number, 10)
        .toString(16);
}

function hex2bin(hex)
{
    var bytes = [], str;

    for(var i=0; i< hex.length-1; i+=2)
        bytes.push(parseInt(hex.substr(i, 2), 16));

    return String.fromCharCode.apply(String, bytes);
}


if (!window.atob) {
    var tableStr = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
    var table = tableStr.split("");

    window.atob = function (base64) {
        if (/(=[^=]+|={3,})$/.test(base64)) throw new Error("String contains an invalid character");
        base64 = base64.replace(/=/g, "");
        var n = base64.length & 3;
        if (n === 1) throw new Error("String contains an invalid character");
        for (var i = 0, j = 0, len = base64.length / 4, bin = []; i < len; ++i) {
            var a = tableStr.indexOf(base64[j++] || "A"), b = tableStr.indexOf(base64[j++] || "A");
            var c = tableStr.indexOf(base64[j++] || "A"), d = tableStr.indexOf(base64[j++] || "A");
            if ((a | b | c | d) < 0) throw new Error("String contains an invalid character");
            bin[bin.length] = ((a << 2) | (b >> 4)) & 255;
            bin[bin.length] = ((b << 4) | (c >> 2)) & 255;
            bin[bin.length] = ((c << 6) | d) & 255;
        };
        return String.fromCharCode.apply(null, bin).substr(0, bin.length + n - 4);
    };

    window.btoa = function (bin) {
        for (var i = 0, j = 0, len = bin.length / 3, base64 = []; i < len; ++i) {
            var a = bin.charCodeAt(j++), b = bin.charCodeAt(j++), c = bin.charCodeAt(j++);
            if ((a | b | c) > 255) throw new Error("String contains an invalid character");
            base64[base64.length] = table[a >> 2] + table[((a << 4) & 63) | (b >> 4)] +
            (isNaN(b) ? "=" : table[((b << 2) & 63) | (c >> 6)]) +
            (isNaN(b + c) ? "=" : table[c & 63]);
        }
        return base64.join("");
    };

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
                var  bin = 1;
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

function bin2hex(s) {
// discuss at: http://phpjs.org/functions/bin2hex/
// original by: Kevin van Zonneveld (http://kevin.vanzonneveld.net)
// bugfixed by: Onno Marsman
// bugfixed by: Linuxworld
// improved by: ntoniazzi (http://phpjs.org/functions/bin2hex:361#comment_177616)
// example 1: bin2hex('Kev');
// returns 1: '4b6576'
// example 2: bin2hex(String.fromCharCode(0x00));
// returns 2: '00'
    var i, l, o = '',
        n;
    s += '';
    for (i = 0, l = s.length; i < l; i++) {
        n = s.charCodeAt(i)
            .toString(16);
        o += n.length < 2 ? '0' + n : n;
    }
    return o;
}


function encode_length(length) {
    if (length <= 0x7F) {
        return bin2hex(String.fromCharCode(length));
    }
    var temp = dechex(length);
    if (temp.length%3==0)
        var temp_hex = '0'+temp;
    else
        var temp_hex = temp;
    temp = hex2bin(temp_hex);
    var new_length = (0x80 | temp.length);
    var new_length_hex = dechex(new_length);
    return (new_length_hex+temp_hex);
}

function make_public_key(n, e) {
    console.log('n='+n);
    console.log('e='+e);
    var n_bin = hex2bin(n);
    var e_bin = hex2bin(e);
    var modulus = '02' + encode_length(n_bin.length) + n;
    var publicExponent = '02' + encode_length(e_bin.length) + e;
    var modulus_bin = hex2bin(modulus);
    var publicExponent_bin = hex2bin(publicExponent);
    var RSAPublicKey = '30' + encode_length(modulus_bin.length + publicExponent_bin.length) + modulus + publicExponent;
    var RSAPublicKey_bin = hex2bin(RSAPublicKey);
    console.log(RSAPublicKey);
    var rsaOID = '300d06092a864886f70d0101010500';
    var rsaOID_bin = hex2bin(rsaOID);
    RSAPublicKey_bin = String.fromCharCode(0) + RSAPublicKey_bin;
    RSAPublicKey_bin = String.fromCharCode(3) + hex2bin(encode_length(RSAPublicKey_bin.length)) + RSAPublicKey_bin;
    RSAPublicKey = bin2hex(RSAPublicKey_bin);
    RSAPublicKey = '30' + encode_length(rsaOID_bin.length + RSAPublicKey_bin.length) + rsaOID + RSAPublicKey;
    console.log(RSAPublicKey);
    return RSAPublicKey;
}


function get_img_refs (i, user_id, urls) {
    if (typeof urls == 'undefined' )
        return 0;
    var image = new Image();
    if (typeof urls[i] != 'undefined' && urls[i]!='' && urls[i]!='0') {
        console.log('TRY '+urls[i]+"/public/"+user_id+"_user_face.jpg"+"\ni="+i);
        image.src = urls[i]+"/public/"+user_id+"_user_face.jpg";
        image.onload = function(){
            console.log('OK '+urls[i]);
            image=null;
            $('.img_'+user_id).css("background", 'url('+urls[i]+'/public/'+user_id+'_user_face.jpg)  50% 50%');
            $('.img_'+user_id).css("background-size", "60px Auto");
        };
        // handle failure
        image.onerror = function(){
            image=null;
            console.log('error '+urls[i]);
            var bg = $('.img_'+user_id).css("background-image");
            if (typeof bg == 'undefined' || bg.length<10)
                get_img_refs (i+1, user_id, urls);
        };
        setTimeout
        (
            function()
            {
                if ( image!=null && (!image.complete || !image.naturalWidth) )
                {
                    var bg = $('.img_'+user_id).css("background-image");
                    image = null;
                    console.log('error');
                    if (typeof bg == 'undefined' || bg.length<10)
                        get_img_refs (i+1, user_id, urls);
                }
            },
            2000
        );
    }
}

function check_form( callback ) {
	var tocheck = {"list": []};
	$(".checkform").remove();
	$(".form-control[check]").each( function(){
		item = $(this);
		var id = item.attr('id');
		if ( item.is(':visible')) {
			var name = $("label[for='"+ id +"']").html();
			if ( typeof name === "undefined" )
				name = '';
			tocheck.list.push( { action: item.attr('check'), id: id,
			                   value: item.val(), label: name } );
   		}
	});
	if ( tocheck.list.length > 0 ) {
		$.post( 'ajaxjson?controllerName=CheckForm', { tocheck: JSON.stringify( tocheck ) }, function( data ) {
			if (data.result || !data.success) {
				if ( callback ) {
					callback( data );
				}
			}
			else {
				data.data = JSON.parse( data.data );
				for (i=0; i<data.data.warnings.length; i++ ) {
					var item = data.data.warnings[i];
					$("#"+item.id).before( '<div class="checkform alert alert-danger alert-dismissable" style="margin-top: 5px">' +
					item.text + '</div>' );
				}
			}
		})
	}
}

function updDcoin() {
	$('.UpdateMessage .alert').html('<img src="/static/img/squares.gif" style="width:20px; margin:0px">');
//	$('.UpdateMessage').prop('disabled', true);

	$.get('ajax?controllerName=UpdateDcoin', function (data) {
		if (typeof data.success !== 'undefined') {
			$('.UpdateMessage .alert').html("Download succeed");
		}
		$('.UpdateMessage .alert').html('complete');
	}, 'JSON');
}

function open_url( obj ) {
	if ( typeof THRUST != "undefined" ) {
		THRUST.remote.send($(obj).attr('href'));
		return false;
	}
	return true;
}
