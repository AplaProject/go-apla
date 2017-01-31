var MenuAPI;
var qDLT = 1000000000000000000;
var g_menuShow = true;
var GKey = {
	init: function () {
		var pass = getCookie('psw');
		var pubKey = localStorage.getItem('PubKey');
		if (pubKey)
			GKey.Public = pubKey;

		if (pass && localStorage.getItem('EncKey')) {
			GKey.decrypt(localStorage.getItem('EncKey'), pass)
		}
		if (localStorage.getItem('Address'))
			GKey.Address = localStorage.getItem('Address');
		var pubKey = localStorage.getItem('PubKey');
		var stateId = localStorage.getItem('StateId');
		if (stateId)
			GKey.StateId = stateId;
		var citizenId = localStorage.getItem('CitizenId');
		if (citizenId)
			GKey.CitizenId = citizenId;
		if (localStorage.getItem('Accounts'))
			GKey.Accounts = JSON.parse(localStorage.getItem('Accounts'));
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

var hist = [['load_template', 'dashboard_default', {}]];
var hist_cur = 0;
var hist_stay = 0;

function hist_push(obj) {
	if (hist_stay != 0) {
		hist_cur += hist_stay;
		hist_stay = 0;
	} else {
		if (hist_cur < hist.length - 1) {
			hist = hist.slice(0, hist_cur + 1);
		}
		hist.push(obj);
		hist_cur = hist.length - 1;
		if (hist.length >= 100) {
			hist.shift();
		}
	}
	if (hist_cur >= 1) {
		$("#hist_back").show();
	} else {
		$("#hist_back").hide();
	}
	if (hist_cur < hist.length - 1) {
		$("#hist_forward").show();
	} else {
		$("#hist_forward").hide();
	}
}

function hist_go(obj) {
	if (obj[0] == 'load_app') {
		load_app(obj[1]);
	} else {
		window[obj[0]](obj[1], obj[2]);
	}
}

function hist_back() {
	if (hist_cur > 0) {
		hist_stay = -1;
		hist_go(hist[hist_cur - 1]);
	}
	return false;
}

function hist_forward() {
	if (hist_cur < hist.length - 1) {
		hist_stay = 1;
		hist_go(hist[hist_cur + 1]);
	}
	return false;
}

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
	/*AllTimer = setTimeout(function () { }, 0);

	for (var i = 0; i < AllTimer; i += 1) {
		if (IgnoreTimer != i) {
			clearTimeout(i);
		}
	}*/
	$(".wrapper").removeClass("map");

	try {
		if (latestTime) {
			clearTimeout(latestTime);
		}
	} catch (err) {
	}
}

function load_page(page, parameters, anchor) {
	//    $('#loader').spin();
	$(".mm-selected").removeClass("mm-selected");
	clearAllTimeouts();
	NProgress.set(1.0);
	$.post("content?page=" + page, parameters ? parameters : {},
		function (data) {
			//            $("#loader").spin(false);
			$(".sweet-overlay, .sweet-alert").remove();
			$('#dl_content').html(data);
			updateLanguage("#dl_content .lang");
			//loadLanguage();
			hist_push(['load_page', page, parameters ? parameters : {}]);
			window.scrollTo(0, 0);
			if (anchor) {
				anchorScroll(anchor);
			}
		}, "html");
}

function clearTempMenu() {
	console.log('clear temp');
	//	$("#mmenu-panel li:first ul").remove();
	//	$("#mmenu-panel li:first a").remove();
	//	$("#ultemporary").remove();
	curMenu = 'main_menu';
}

var latestMenu = '';
var curMenu = 'main_menu';

function ajaxMenu(page, parameters) {
	$.ajax({
		url: 'ajax?controllerName=ajaxGetMenuHtml&page=' + page,
		type: 'POST',
		data: parameters ? parameters : {},
		success: function (data) {
			// linked menu
			var amenuname = data.match(/<!--#([\w_\d]*)#-->/) || [""];
			if (amenuname.length > 1)
				menuname = amenuname[1];
			console.log('Main', menuname, curMenu);
			if (menuname == 'main_menu') {
				if (curMenu != menuname) {
					curMenu = menuname;
					MenuAPI.openPanel($("#mmenu-panel"));
				}
				return;
			}
			if (curMenu == menuname)
				return;
			curMenu = menuname;


			var name = 'temporary';
			var aname = data.match(/<!--([\w_\d]*)-->/) || [""];
			if (aname.length > 1) {
				name = aname[1];
			} else {
				//				$("#ultemporary").remove();
			}
			/*			if (curMenu == name)
							return;*/
			/*				//				$("#mmenu-panel ul").remove();
							$("#mmenu-panel li:first").append('<ul id="ul' + name + '">' + data + '</ul>');
							updateLanguage($("#ul" + name + ' .lang'));
							MenuAPI.initPanels($("#ul" + name));
							MenuAPI.openPanel($("#ul" + name));
			*/
			console.log('Menu', menuname, name, latestMenu, curMenu);
			if (latestMenu != '' && latestMenu != name) {
				//					MenuAPI.openPanel($("#mmenu-panel"));
				//					MenuAPI.setSelected($("#li" + latestMenu), true);
				$("#ul" + latestMenu).remove();
				$("#li" + latestMenu + ' .mm-next').remove();
				//					MenuAPI.update();
			}
			console.log('Menu 2', latestMenu, name);
			if (latestMenu != name || name == 'temporary') {
				latestMenu = name;
				if (name != 'temporary') {
					$("#li" + name + " ul").remove();
					$("#li" + name).append('<ul id="ul' + name + '">' + data + '</ul>');
				} else {
					$("#ultemporary").remove();
					console.log('Dynamic');
					$("#mmenu-panel li:first").append('<ul id="ul' + name + '">' + data + '</ul>');
				}
				updateLanguage($("#ul" + name + ' .lang'));
				MenuAPI.initPanels($("#ul" + name));
			}
			MenuAPI.openPanel($("#ul" + name));
			$("#li" + name + ' .mm-next').remove();
			$(".mm-selected").removeClass("mm-selected");
			MenuAPI.setSelected($("#ul" + name + " #li" + page), true);
			if (name == 'temporary') {
				$("#mmenu-panel li:first a").remove();
			}
			var bname = data.match(/<!--([\w_\d ]*)=([\w_\d '\(\)]*)-->/) || [""];
			if (bname.length > 2) {
				$(".mm-navbar-top .mm-title").html(bname[1]);
				if (bname[2].length > 0) {
					$(".mm-navbar-top a").attr('href', '');
					$(".mm-navbar-top a").attr('onclick', bname[2] + ';  return false;');
				}
			} else {
				if (name == 'temporary') {
					var title = $("#mmenu-panel .mm-navbar a").html();
					$(".mm-navbar-top a").attr('onclick', '');
					$(".mm-navbar-top .mm-title").html(title);
					$(".mm-navbar-top a").attr('href', '#mmenu-panel');
					$(".mm-navbar-top a").attr('onclick', 'clearTempMenu()');
				}
				else
					$(".mm-navbar-top a").attr('onclick', 'clearTempMenu()');
			}
		}

		/*		    if (data.length == 0) {
						return
					}
					var aname = data.match(/<!--([\w_\d]*)-->/) || [""];
					var name = 'temporary';
					if (aname.length > 1) {
						name = aname[1];
					} else {
					}
					if (latestMenu != '' && latestMenu != name) {
						$("#ul" + latestMenu).remove();
						$("#li" + latestMenu + ' .mm-next').remove();
						//					MenuAPI.update();
					}
					console.log('Menu', latestMenu, name);
					if (latestMenu != name || latestMenu == 'temporary') {
						latestMenu = name;
						if (name != 'temporary') {
							$("#li" + name + " ul").remove();
							$("#li" + name).append('<ul id="ul' + name + '">' + data + '</ul>');
						} else {
							console.log('Dynamic');
							$("#ultemporary").remove();
							$("#mmenu-panel li:first").append('<ul id="ul' + name + '">' + data + '</ul>');
						}
						updateLanguage($("#ul" + name + ' .lang'));
						MenuAPI.initPanels($("#ul" + name));
					}
					MenuAPI.openPanel($("#ul" + name));
					$("#li" + name + ' .mm-next').remove();
					$(".mm-selected").removeClass("mm-selected");
					MenuAPI.setSelected($("#ul" + name + " #li" + page), true);
					if (name == 'temporary') {
						$("#mmenu-panel li:first a").remove();
					}
					var bname = data.match(/<!--([\w_\d ]*)=([\w_\d '\(\)]*)-->/) || [""];
					if (bname.length > 2) {
						$(".mm-navbar-top .mm-title").html(bname[1]);
						if (bname[2].length > 0) {
							$(".mm-navbar-top a").attr('href', '');
							$(".mm-navbar-top a").attr('onclick', bname[2] + ';return false;');
						}
					} else {
						if (name == 'temporary') {
							var title = $("#mmenu-panel .mm-navbar a").html();
							$(".mm-navbar-top a").attr('onclick', '');
							$(".mm-navbar-top .mm-title").html(title);
							$(".mm-navbar-top a").attr('href', '#mmenu-panel');
							//	$(".mm-navbar-top a").attr('onclick', 'clearTempMenu()');
						}
						else
							$(".mm-navbar-top a").attr('onclick', '');
					}
				}*/
	});
}

function load_template(page, parameters, anchor) {
	clearAllTimeouts();
	NProgress.set(1.0);
	$.post("template?page=" + page, parameters ? parameters : {},
		function (data) {
			if (data == '') {
				load_page('newPage', { global: 0, name: page });
				return;
			}
			$(".sweet-overlay, .sweet-alert").remove();
			$('#dl_content').html(data);
			updateLanguage("#dl_content .lang");
			//loadLanguage();
			hist_push(['load_template', page, parameters ? parameters : {}]);
			window.scrollTo(0, 0);
			ajaxMenu(page, parameters);
			
			if (anchor) {
				anchorScroll(anchor);
			}
		}, "html");
}

function load_app(page) {
	clearAllTimeouts();
	NProgress.set(1.0);
	$.post("app?page=" + page, {},
		function (data) {
			$(".sweet-overlay, .sweet-alert").remove();
			$('#dl_content').html(data);
			updateLanguage("#dl_content .lang");
			//loadLanguage();
			hist_push(['load_app', page]);
			window.scrollTo(0, 0);
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

function defaultConfirm() {
	return true;
}

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
			if (!obj.hasClass("modal-content")) {
				Alert("", returnLang("copied_clipboard"), "notification:success", defaultConfirm);
			} else {
				Alert(returnLang("copied_clipboard"), "", "success", defaultConfirm);
			}
		}
		$(elem).addClass("copied");
		setTimeout(function () {
			$(elem).removeClass("copied");
		}, 3000);
	});
	clipboard.on('error', function (e) {
		if (!obj.hasClass("modal-content")) {
			Alert("", returnLang("error_copying_clipboard"), "notification:danger", defaultConfirm);
		} else {
			Alert(returnLang("error_copying_clipboard"), "", "error", defaultConfirm);
		}
	});
}

function Alert(title, text, type, Confirm) {
	if (obj) {
		var timer = null;
		var view = type.split(":");

		if (view[0] == "notification" && !obj.hasClass("modal-content")) {
			type = view[1] ? view[1] : "success";
			timer = 1500;

			$.notify({
				message: text,
				status: type,
				timeout: timer
			});

			if (Confirm) {
				if (Confirm == false) {
					return false;
				} else {
					Confirm();
				}
			}
			if (Confirm != false) {
				setTimeout(function () {
					obj.removeClass("whirl standard");
				}, 0)
			}
		} else {
			var color;
			var btnText = "OK";
			var btnShow = true;
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
					btnText = returnLang("copy_text_error_clipboard");
				}
			} else if (type == "warning") {
				color = "#ff902b";
			} else if (type == "timeout") {
				type = "success";
				color = "#23b7e5";
				btnShow = false;
				timer = 3000;
			} else if (view[0] == "notification") {
				type = view[1] ? view[1] : "success";
				if (type == "success") {
					color = "#23b7e5";
				} else if (type == "danger") {
					type = "error";
					color = "#f05050";
					if (text.toLowerCase().indexOf("[error]") != -1) {
						btnText = returnLang("copy_text_error_clipboard");
					}
				} else if (type == "warning") {
					color = "#ff902b";
				}
			} else {
				color = "#c1c1c1";
			}

			$(".sweet-alert").appendTo($("body"));

			swal({
				title: title,
				text: text,
				timer: timer,
				allowEscapeKey: false,
				type: type,
				html: true,
				confirmButtonColor: color,
				confirmButtonText: btnText,
				showConfirmButton: btnShow
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

				if (timer) {
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
					swal.close();
				}
			});

			if (bh > oh) {
				$(".sweet-alert").appendTo(obj);
			}
		}
	}
}

function preloader_hide() {
	$(".sk-cube-grid").remove();
	$(".whirl").removeClass('whirl');
}

function preloader(elem) {
	obj = $("#" + elem.id).parents("[data-sweet-alert]");
	if (!obj.find(".sk-cube-grid").length) {
		obj.append('<div class="sk-cube-grid"><div class="sk-cube sk-cube1"></div><div class="sk-cube sk-cube2"></div><div class="sk-cube sk-cube3"></div><div class="sk-cube sk-cube4"></div><div class="sk-cube sk-cube5"></div><div class="sk-cube sk-cube6"></div><div class="sk-cube sk-cube7"></div><div class="sk-cube sk-cube8"></div><div class="sk-cube sk-cube9"></div></div>');
	}
}

function dl_navigate(page, parameters, anchor) {
	if (page.substr(0, 2) == 'ul') {
		return;
	}
	var json = JSON.stringify(parameters);
	//$('#loader').spin();
	clearAllTimeouts();
	NProgress.set(1.0);
	$.post("content?controllerHTML=" + page, { tpl_name: page, parameters: json },
		function (data) {
			//$("#loader").spin(false);
			$(".sweet-overlay, .sweet-alert").remove();
			$('#dl_content').html(data);
			updateLanguage("#dl_content .lang");
			//loadLanguage();
			hist_push(['dl_navigate', page, parameters]);
			/*if ( parameters && parameters.hasOwnProperty("lang")) {
				if ( page[0] == 'E' )
					load_emenu();
				else
					load_menu();
			}*/
			window.scrollTo(0, 0);
			
			if (anchor) {
				anchorScroll(anchor);
			}
		}, "html");
}

function load_menu(lang, submenu) {
	if (g_menuShow) {
		parametersJson = "";
		if (typeof lang != 'undefined') {
			parametersJson: '{"lang":"1"}'
		}
		$("#dl_menu").load("content?page=menu", { parameters: parametersJson }, function () {
			updateLanguage("#dl_menu .lang");
			if (typeof submenu === "string")
				ajaxMenu(submenu);
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
	unique = '';
	if (typeof (type) === 'number') {
		unique = type.toString();
		type = 'sign';
	}
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
		var forsignature = $("#for-signature" + unique).val();
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
		$("#signature1" + unique).val(signature);
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
	var i = 0;
	
	if (typeof data.error != "undefined" && data.error.length > 0) {
		Alert(returnLang("error"), data.error, "error");
	} else if (data.hash == "undefined") {
		Alert(returnLang("error"), data.result, "error");
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
						if (txStatus.error[0] == '!') {
							re = /!(.*)\(parser/i;
							found = txStatus.error.match(re);
							Alert(returnLang("warning"), (found && found.length > 1 ? found[1] : txStatus.error.substr(1)), "warning");
						} else if (txStatus.error[0] == '*') {
							re = /\*(.*)\(parser/i;
							found = txStatus.error.match(re);
							Alert(returnLang("info"), (found && found.length > 1 ? found[1] : txStatus.error.substr(1)), "info");
						} else
							Alert(returnLang("error"), txStatus.error, "error");
						clearInterval(interval);
					} else {
						clearInterval(interval);
						block_explorer = 'block_explorer';
						if (skipsuccess) {
							ReadyFunction(txStatus.success);
						} else {
							//Alert(returnLang("success"), 'Imprinted in blockchain. Block <a href="#" onclick="load_page(' + block_explorer + ', {blockId: ' + txStatus.success + '});">' + txStatus.success + '</a>',
							Alert(returnLang("success"), returnLang("imprinted_blockchain") + ' ' + txStatus.success + '',
								typeof data.type_success === "string" ? data.type_success : 'notification', ReadyFunction);
						}
					}
				},
				error: function (xhr, status, error) {
					clearInterval(interval);
					Alert(returnLang("error"), error, "error");
				},
			});
			
			i += 1;
			
			if (i >= 6) {
				clearInterval(interval);
				Alert(returnLang("error"), returnLang("cannot_connect_server"), "notification:danger");
			}
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
			Alert(returnLang("warning"), "Please, choose image!", "warning", false);
		}
	} else {
		Alert(returnLang("warning"), "Please, crop the photo!", "warning", false);
	}
}

function openBlockDetailPopup(id) {
	$("#dl_modal").load("content?page=block_explorer", { modal: 1, blockId: id }, function () {
		var modal = $("#modal_block_detail");

		modal.modal("show");
	});
}

function formatCode() {
	var source = editor.getValue();
	var output = js_beautify(source, {
		'indent_size': 1,
		'indent_char': '\t'
	});
	var cursor = editor.getCursorPosition();
	editor.setValue(output, -1);
	editor.moveCursorToPosition(cursor);
	editor.focus();
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

	for (i = 0, len = str.length; i < len; i += n) {
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

function MoneyDigit(value, dig) {
	var money = value.replace(' ', '');
	var digit = parseInt(dig, 10);
	if (digit > 0) {
		off = money.indexOf('.');
		if (off < 0) {
			money = money + '0'.repeat(digit);
		} else {
			var cents = money.substr(off + 1);
			if (cents.length > digit) {
				money = money.substr(0, off) + cents.substr(0, digit);
			} else if (cents.length < digit) {
				money = money.substr(0, off) + cents + '0'.repeat(digit - cents.length);
			}
		}
	}
	return money.replace('.', '');
}

function loadLanguage() {
	//localStorage.removeItem('EGAAS_LANG');
	var userLang = navigator.language || navigator.userLanguage;
	var lang = localStorage.getItem('EGAAS_LANG');

	if (lang === null) {
		lang = userLang.substring(0, 2);
		if (lang != "nl") {
			lang = "gb";
		}
	}

	localStorage.setItem('EGAAS_LANG', lang);

	changeLanguage(lang);
}

function updateLanguage(classes) {

	if (typeof Lang === "undefined")
		loadLanguage();
	else
		$(classes).each(function (obj) {
			var data = $(this).attr('lang-id');
			$(this).html(Lang[data]);
		});
	$("#langflag").attr('class', 'flag ' + localStorage.getItem('EGAAS_LANG'));
}

function loadjs(filename) {
	var fileref = document.createElement('script')
	fileref.setAttribute("type", "text/javascript")
	fileref.setAttribute("src", filename)
	if (typeof fileref != "undefined")
		document.getElementsByTagName("head")[0].appendChild(fileref)
}

function changeLanguage(lang) {
	loadjs('static/lang/' + lang + '.js');
	/*	setTimeout(function () {
			updateLanguage('.lang');
		}, 500);*/
	$("#langflag").attr('class', 'flag ' + lang);
	localStorage.setItem('EGAAS_LANG', lang);
	//$("select").trigger("change.select2");
}

function returnLang(data) {
	if (typeof Lang === 'undefined')
		return '';
	return Lang[data];
}

function anchorScroll(anchor) {
	var top = "#" + anchor;
	setTimeout(function () {
		$.scrollTo(top, 300, {easing:'linear', onAfter: function() {
			// Можно добавить что-то после скроллинга
		}});
	}, 1000);
}

function InitMobileHead() {
	var head = $(".content-wrapper .content-heading").clone().children().remove().end().text();
	$(".topnavbar-wrapper .content-heading").text(head);
}

function InitMobileTable() {
	var table = $("[data-role='table']");
	table.data('mode', 'reflow').addClass("ui-responsive");

	if (table.length) {
		//console.log('load table');
		table.each(function () {
			var _this = $(this);
			_this.find("tbody tr").each(function () {
				var td = $(this).find("td");
				var title = $(this).find("td:first");

				if (!td.find(".ui-table-td").length) {
					td.wrapInner("<div class='ui-table-td'></div>");
				}
				title.addClass("ui-table-title");
			});
			if (_this.hasClass("ui-table")) {
				_this.closest("table").table("refresh").trigger("create");
				//console.log('reload table');
			} else {
				_this.table();
			}
		});

		$(".column_type").each(function () {
			var id = $(this);
			var val = id.val();
			if (val === "text") {
				id.parent().parent().parent().find(".index").prop("disabled", true);
			} else {
				id.parent().parent().parent().find(".index").prop("disabled", false);
			}
		});
	}
}

var observeDOM = (function () {
	var MutationObserver = window.MutationObserver || window.WebKitMutationObserver,
		eventListenerSupported = window.addEventListener;

	return function (content, callback) {
		if (MutationObserver) {
			// define a new observer
			var obs = new MutationObserver(function (mutations, observer) {
				if (mutations[0].addedNodes.length || mutations[0].removedNodes.length)
					callback();
			});
			// have the observer observe foo for changes in children
			obs.observe(content, { childList: true, subtree: true });
		}
		else if (eventListenerSupported) {
			content.addEventListener('DOMNodeInserted', callback, false);
			content.addEventListener('DOMNodeRemoved', callback, false);
		}
	}
})();

observeDOM(document.getElementById('dl_content'), function () {
	InitMobileHead();
	InitMobileTable();
});

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