var qDLT = 1000000000000000000;
var g_menuShow = true;
var GKey = {
	init: function () {
		var pass = getCookie('psw');
		var pubKey = localStorage.getItem('PubKey');
		if (pubKey)
			this.Public = pubKey;

		if (pass && localStorage.getItem('EncKey')) {
			this.decrypt(localStorage.getItem('EncKey'), pass)
		}
		if (localStorage.getItem('Address'))
			this.Address = localStorage.getItem('Address');
		var pubKey = localStorage.getItem('PubKey');
		var stateId = localStorage.getItem('StateId');
		if (stateId)
			this.StateId = stateId;
		var citizenId = localStorage.getItem('CitizenId');
		if (citizenId)
			this.CitizenId = citizenId;
		if (localStorage.getItem('Accounts'))
			this.Accounts = JSON.parse(localStorage.getItem('Accounts'));
	},
	add: function (address) {
		localStorage.setItem('Address', address);
		GKey.Address = address;
		var data = {
			EncKey: localStorage.getItem('EncKey'),
			//			Encrypt: localStorage.getItem('Encrypt'),
			Public: GKey.Public,
			Address: address,
			StateId: GKey.StateId,
			CitizenId: GKey.CitizenId,
		}
		for (i = 0; i < this.Accounts.length; i++) {
			if (this.Accounts[i].Address == address) {
				this.Accounts[i] = data;
				break;
			}
		}
		if (i >= this.Accounts.length)
			this.Accounts.push(data);
		localStorage.setItem('Accounts', JSON.stringify(this.Accounts));
		//		if (thrust)
		//			$.post("ajax?json=ajax_storage",{accounts: localStorage.getItem('Accounts')});
		if (typeof THRUST != "undefined")
			THRUST.remote.send(localStorage.getItem('Accounts'));

	},
	clear: function () {
		//		localStorage.removeItem('PubKey');
		localStorage.removeItem('EncKey');
		localStorage.removeItem('Address');
		this.Address = '';
		this.StateId = '';
		this.CitizenId = '';
		deleteCookie('psw');
	},
	decrypt: function (encKey, pass) {
		var decrypted = CryptoJS.AES.decrypt(encKey, pass).toString(CryptoJS.enc.Hex);
		var prvkey = '';
		for (i = 0; i < decrypted.length; i += 2) {
			var num = parseInt(decrypted.substr(i, 2), 16);
			prvkey += String.fromCharCode(num);
		}
		if (this.verify(prvkey, this.Public)) {
			this.Private = prvkey;
			this.Password = pass;
			return true;
		}
		return false;
	},
	save: function (seed) {
		localStorage.setItem('EncKey', CryptoJS.AES.encrypt(this.Private, this.Password));
		localStorage.setItem('PubKey', GKey.Public);
		localStorage.setItem('CitizenId', GKey.CitizenId);
		localStorage.setItem('StateId', GKey.StateId);
		if (seed)
			localStorage.setItem('Encrypt', CryptoJS.AES.encrypt(seed, this.Password));
		setCookie('psw', this.Password);
	},
	sign: function (msg, prvkey) {
		if (!prvkey) {
			prvkey = this.Private
		}
		var sig = new KJUR.crypto.Signature({ "alg": this.SignAlg });
		sig.initSign({ 'ecprvhex': prvkey, 'eccurvename': this.Curve });
		sig.updateString(msg);
		return sig.sign();
	},
	verify: function (prvkey, pubkey) {
		var msg = 'test';
		var sigval = this.sign(msg, prvkey);
		var siga = new KJUR.crypto.Signature({ "alg": this.SignAlg, "prov": "cryptojs/jsrsa" });
		siga.initVerifyByPublicKey({ 'ecpubhex': pubkey, 'eccurvename': this.Curve });
		siga.updateString(msg);
		return siga.verify(sigval);
	},
	SignAlg: 'SHA256withECDSA',
	Curve: 'secp256r1',
	Accounts: [],
	Password: '',
	Private: '',
	Public: '',
	Address: '',
	StateId: '',
	CitizenId: ''
}

GKey.init();

function getCookie(name) {
	var matches = document.cookie.match(new RegExp(
		"(?:^|; )" + name.replace(/([\.$?*|{}\(\)\[\]\\\/\+^])/g, '\\$1') + "=([^;]*)"
	));
	return matches ? decodeURIComponent(matches[1]) : undefined;
}

function deleteCookie(name) {
	setCookie(name, "", {
		expires: -1
	})
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

function logout() {
	GKey.clear();
	$.get("ajax?controllerName=logout",
		function () {
			window.location.href = "/";
		});

	return false;
}

var AllTimer;
var IgnoreTimer;

function clearAllTimeouts() {
	AllTimer = setTimeout(function () { }, 0);

	for (var i = 0; i < AllTimer; i += 1) {
		if (IgnoreTimer != i) {
			clearTimeout(i);
		}
	}

	$(".wrapper").removeClass("map");
}

function load_page(page, parameters) {
	//    $('#loader').spin();
	clearAllTimeouts();
	NProgress.set(1.0);
	$.post("content?page=" + page, parameters ? parameters : {},
		function (data) {
			//            $("#loader").spin(false);
			$(".sweet-overlay, .sweet-alert").remove();
			$('#dl_content').html(data);
			window.scrollTo(0, 0);
			if ($(".sidebar-collapse").is(":visible") && $(".navbar-toggle").is(":visible"))
				$('.sidebar-collapse').collapse('toggle');
		}, "html");
}


function load_template(page, parameters) {
	clearAllTimeouts();
	NProgress.set(1.0);
	$.post("template?page=" + page, parameters ? parameters : {},
		function (data) {
			$(".sweet-overlay, .sweet-alert").remove();
			$('#dl_content').html(data);
			window.scrollTo(0, 0);
			if ($(".sidebar-collapse").is(":visible") && $(".navbar-toggle").is(":visible")) {
				$('.sidebar-collapse').collapse('toggle');
			}
			console.log(page);
			$.ajax({
				url: 'ajax?controllerName=ajaxGetMenuHtml&page=' + page,
				type: 'POST',
				success: function (data) {
					console.log(data);
					var li = $("#dc li:first").html();
					$("#dc").html('<li class="sidebar-subnav-header">' + li + '</li>' + data);
					$("#dc li:first").next().addClass("active");
				}
			});
		}, "html");
}

function load_app(page) {
	clearAllTimeouts();
	NProgress.set(1.0);
	$.post("app?page=" + page, {},
		function (data) {
			$(".sweet-overlay, .sweet-alert").remove();
			$('#dl_content').html(data);
			window.scrollTo(0, 0);
			if ($(".sidebar-collapse").is(":visible") && $(".navbar-toggle").is(":visible"))
				$('.sidebar-collapse').collapse('toggle');
		}, "html");
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

function Notify(message, options) {
	var btn_notify = $("#notify");
	btn_notify.data("message", message);
	btn_notify.data("options", options);
	btn_notify.click();
}

var clipboard;

function CopyToClipboard(elem, text) {
	if (clipboard) {
		clipboard.destroy();
	}
	clipboard = new Clipboard(elem);

	if (text) {
		$(elem).attr("data-clipboard-text", text);
	}

	clipboard.on('success', function (e) {
		e.clearSelection();
		if (text) {
			$(elem).attr("data-clipboard-text", "");
		} else {
			Alert("Copied to clipboard", "", "success");
		}
		$(elem).addClass("copied");
		setTimeout(function(){
			$(elem).removeClass("copied");
		}, 3000)
	});
	clipboard.on('error', function (e) {
		Alert("Error copying to clipboard", "", "error");
	});
}

function Alert(title, text, type, Confirm) {
	if (obj) {
		var color;
		var btnText = "OK";
		var id = obj.parents(".modal").attr("id");
		var bh = window.innerHeight - 170;
		var oh = obj.height();
		var minHeight = obj.css("min-height");
		obj.css({ "position": "relative", "min-height": "300px" });

		if (type == "success") {
			color = "#23b7e5";
		} else if (type == "error") {
			color = "#f05050";
			if (text.toLowerCase().indexOf("[error]") != -1) {
				btnText = "Copy text error to clipboard";
			}
		} else if (type == "warning") {
			color = "#ff902b";
		} else {
			color = "#c1c1c1";
		}
		
		$(".sweet-overlay, .sweet-alert").appendTo($("body"));

		swal({
			title: title,
			text: text,
			allowEscapeKey: false,
			type: type,
			html: true,
			confirmButtonColor: color,
			confirmButtonText: btnText
		}, function (isConfirm) {
			if (text.toLowerCase().indexOf("[error]") != -1) {
				CopyToClipboard(".sweet-alert .confirm", text);
			}
			if (isConfirm) {
				if (Confirm) {
					if (Confirm == false) {
						return false;
					} else {
						Confirm();
					}
				}
				if (Confirm != false) {
					obj.css({ "min-height": minHeight }).removeClass("whirl standard");
					minHeight = null;
					$("#" + id).modal("hide");
				}
			}
		});
		
		if (bh > oh) {
			$(".sweet-overlay, .sweet-alert").appendTo(obj);
		}
	}
}

function preloader(elem) {
	obj = $("#" + elem.id).parents("[data-sweet-alert]");
	if (!obj.find(".sk-cube-grid").length) {
		obj.append('<div class="sk-cube-grid"><div class="sk-cube sk-cube1"></div><div class="sk-cube sk-cube2"></div><div class="sk-cube sk-cube3"></div><div class="sk-cube sk-cube4"></div><div class="sk-cube sk-cube5"></div><div class="sk-cube sk-cube6"></div><div class="sk-cube sk-cube7"></div><div class="sk-cube sk-cube8"></div><div class="sk-cube sk-cube9"></div></div>');
	}
}

function dl_navigate(page, parameters) {
	var json = JSON.stringify(parameters);
	//$('#loader').spin();
	clearAllTimeouts();
	NProgress.set(1.0);
	$.post("content?controllerHTML=" + page, { tpl_name: page, parameters: json },
		function (data) {
			//$("#loader").spin(false);
			$(".sweet-overlay, .sweet-alert").remove();
			$('#dl_content').html(data);
			/*if ( parameters && parameters.hasOwnProperty("lang")) {
				if ( page[0] == 'E' )
					load_emenu();
				else
					load_menu();
			}*/
			window.scrollTo(0, 0);
		}, "html");
}

function load_menu(lang) {
	if (g_menuShow) {
		parametersJson = "";
		if (typeof lang != 'undefined') {
			parametersJson: '{"lang":"1"}'
		}
		$("#dl_menu").load("content?page=menu", { parameters: parametersJson }, function () {
		});
	} else {
		$("#dl_menu").html('');
	}
}

function MenuReload() {
	load_menu();
}

function login_ok(result) {
	g_menuShow = true;
	load_menu();

	setTimeout(function () {
		if (result) {
			$("#dl_content").load("content", { tpl_name: 'home' }, function () {
				NProgressStart.done();
			});
		}
	}, 100);
}

function doSign_(type) {

	if (typeof (type) === 'undefined') type = 'sign';

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

	if (!GKey.Private) {
		$("#modal_alert").html('<div id="alertModalPull" class="alert alert-danger alert-dismissable"><button type="button" class="close" data-dismiss="alert" aria-hidden="true">×</button><p>' + $('#incorrect_key_or_password').val() + '</p></div>');
		//$("#loader").spin(false);
		return false;
	}
	if (type == 'sign') {
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

	var signature;
	console.log('forsignature=' + forsignature);
	if (forsignature) {
		signature = GKey.sign(forsignature);
	} else {
		return;
	}
	if (SIGN_LOGIN) {

		console.log('SIGN_LOGIN');

		//$("#loader").spin();
		if (key) {
			var privKey = "";
			if ($('#exchangeTemplate').val() == "1") {
				var check_url = 'ajax?controllerName=ECheckSign'
			} else {
				var check_url = 'ajax?controllerName=check_sign'
			}
			// шлем подпись на сервер на проверку
			$.post(check_url, {
				'signature': signature,
				'private_key': privKey,
				'forsignature': forsignature,
			}, function (data) {
				// залогинились
				console.log("data.result: ", data.result);
				login_ok(data.result);

			}, 'JSON'
			);
		}
		else {

			hash_pass = hex_sha256(hex_sha256(pass));
			// шлем хэш пароля на проверку и получаем приватный ключ
			$.post('ajax?controllerName=check_pass', {
				'hash_pass': hash_pass
			}, function (data) {
				// залогинились
				login_ok(data.result);

				$("#modal_key").val(data.key);
				$("#key").text(data.key);
				//alert(data.key);

			}, 'JSON'
			);

		}

		//$("#loader").spin(false);

	}
	else {
		console.log('Signature', signature);
		$("#signature1").val(signature);
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
	image.onload = function () {

		$('#canvas_key').attr('width', this.width);
		$('#canvas_key').attr('height', this.height);
		var c = document.getElementById("canvas_key");
		var ctx = c.getContext("2d");

		ctx.drawImage(image, 0, 0);

		// вначале прочитаем инфу, где искать rsa-ключ (64 пиксла = 64 бита = 8 байт = 4 числа = x,y,w,h)
		var count_bits = 0;
		var byte = '';
		var rsa_search_params = [];
		for (var x = 0; x < 64; x++) {
			var Pixel = ctx.getImageData(x, 0, 1, 1);
			//console.log(x+' '+y+' / '+Pixel.data[0]+' '+Pixel.data[1]+' '+Pixel.data[2]);
			if (Pixel.data[0] > 100)
				var bin = 1;
			else
				var bin = 0;
			byte = byte + '' + bin;
			count_bits = count_bits + 1;
			if (count_bits == 16) {
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
		for (var y = rsa_search_params[1]; y < (Number(rsa_search_params[1]) + Number(rsa_search_params[3])); y++) {
			for (var x = rsa_search_params[0]; x < (Number(rsa_search_params[0]) + Number(rsa_search_params[2]) - 1); x++) {
				var Pixel = ctx.getImageData(x, y, 1, 1);
				//console.log(x+' '+y+' / '+Pixel.data[0]+' '+Pixel.data[1]+' '+Pixel.data[2]);
				if (Pixel.data[0] > 100)
					var bin = 1;
				else
					var bin = 0;
				byte = byte + '' + bin;
				count_bits = count_bits + 1;
				if (count_bits == 8) {
					hex_byte = strpadleft(base_convert(byte, 2, 16));
					//console.log(byte+'='+hex_byte);
					hex = hex + '' + hex_byte;
					count_bits = 0;
					byte = '';
				}
			}
		}
		hex = hex.split('00000000');
		console.log(hex);
		var key = hexToBase64(hex[0]);
		console.log(key);
		$('#' + key_id).val(key);
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

function unixtime(target) {
	if (!target) {
		target = ".unixtime";
	}
	if ($(target).length) {
		$(target).each(function () {
			var time_val = $(this).text();
			if (time_val) {
				var time = Number($(this).text() + '000');
                /*var d = new Date(time);
                $(this).text(d);*/
				var d = new Date();
				d.setTime(time);
				$(this).text(d.toLocaleString());
			}
		});
	}
}

function send_to_net_success(data, ReadyFunction, skipsuccess) {
	if (typeof data.error != "undefined" && data.error.length > 0) {
		Alert("Error", data.error, "error");
	} else if (data.hash == "undefined") {
		Alert("Error", data.result, "error");
	} else {
		interval = setInterval(function () {
			$.ajax({
				type: 'POST',
				url: 'ajax?controllerName=txStatus',
				data: {
					'hash': data.hash
				},
				dataType: 'json',
				crossDomain: true,
				success: function (txStatus) {

					console.log("txStatus", txStatus);

					if (typeof txStatus.wait != "undefined") {
						console.log("txStatus", txStatus);
					} else if (typeof txStatus.error != "undefined") {
						Alert("Error", txStatus.error, "error");
						clearInterval(interval);
					} else {
						clearInterval(interval);
						block_explorer = 'block_explorer';
						if (skipsuccess) {
							ReadyFunction(txStatus.success);
						} else {
							Alert('Success', 'Imprinted in blockchain. Block <a href="#" onclick="load_page(' + block_explorer + ', {blockId: ' + txStatus.success + '});">' + txStatus.success + '</a>', 'success', ReadyFunction);
						}
					}
				},
				error: function (xhr, status, error) {
					clearInterval(interval);
					Alert("Error", error, "error");
				},
			});
		}, 1000)
	}
}

function selectboxState(data) {
	for (var i in data) {
		selectbox.append('<option value="' + i + '" data-id="' + i + '" data-flag="' + data[i].state_flag + '">' + data[i].state_name + '</option>');
	}

	selectbox.select2({
		minimumResultsForSearch: 10,
		templateResult: formatState,
		templateSelection: formatState,
		theme: 'bootstrap'
	});

	selectbox.val(selectbox.find("option:first-child").val()).trigger('change');
};

function formatState(state) {
	if (!state.id) { return state.text; }
	var $state = $(
		'<span class="virtual state_' + state.id + '">' +
		'<i style="background-image:url(' + selectbox.find("option[value=" + state.id + "]").attr("data-flag") + ');"></i>' +
		state.text +
		'</span>'
	);
	return $state;
};

var newImage;
var newImageData;
var PhotoRatio;
var PhotoWidth;
var PhotoHeight;

function openImageEditor(img, container, ratio, width, height) {
	newImage = $("#" + img);
	newImageData = $("#" + container);
	PhotoRatio = ratio.split('/');
	PhotoRatio = PhotoRatio[0] / PhotoRatio[1];
	PhotoWidth = width;
	PhotoHeight = height;

	$("#dl_modal").load("content?controllerHTML=modal_avatar", {}, function () {
		var modal = $("#modal_avatar");

		modal.modal("show");
	});
}

function saveImage() {
	var el = $("#photoEditor #cropped");
	var pts = $("#photoEditor img").length;
	if (!el.hasClass("cropper-hidden")) {
		if (pts > 0) {
			var img = el.attr("src");
			newImage.attr("src", img);
			newImageData.val(img);
			$("#modal_avatar").modal("hide");
		} else {
			Alert("Warning", "Please, choose image!", "warning", false);
		}
	} else {
		Alert("Warning", "Please, crop the photo!", "warning", false);
	}
}

var tagsToReplace = {
	'&': '&amp;',
	'<': '&lt;',
	'>': '&gt;'
};

function replaceTag(tag) {
	return tagsToReplace[tag] || tag;
}

function safe_tags_replace(str) {
	return str.replace(/[&<>]/g, replaceTag);
}

function chunk(str, n) {
	var ret = [];
	var i;
	var len;
	
	for(i = 0, len = str.length; i < len; i += n) {
	   ret.push(str.substr(i, n))
	}
	
	return ret;
}

function FormValidate(form, input, btn) {
	var i = 0;

	form.find("." + input + ":visible").each(function () {
		var val = $(this).val();
		if (val == "") {
			i += 1;
		}
	});

	if (i == 0) {
		btn.prop("disabled", false);
	} else {
		btn.prop("disabled", true);
	}
}

function Validate(form, input, btn) {
	var form = $("#" + form);
	var btn = $("#" + btn);

	FormValidate(form, input, btn);

	form.on('input', function () {
		FormValidate(form, input, btn);
	})
}

$(document).on('keydown', function (e) {
	if (e.keyCode == 13 && $(".keyCode_13:visible").length) {
		if (!$(".select2-container--focus").length) {
			if (!$(".sweet-alert").is(":visible")) {
				$(".submit:not(:disabled)").click();
			} else {
				$(".keyCode_13:visible").find(".sweet-alert:visible .confirm").click();
				$("[data-sweet-alert]").removeClass("whirl standard");
			}
			return false;
		}
	}
});

jQuery.os = { name: (/(win|mac|linux|sunos|solaris|iphone|ipad)/.exec(navigator.platform.toLowerCase()) || [u])[0].replace('sunos', 'solaris') };
if (jQuery.os.name === "mac" || jQuery.os.name === "iphone" || jQuery.os.name === "ipad") {
	$("body").addClass("macfix");
}
if (jQuery.os.name === "linux") {
	$("body").addClass("androidfix");
}