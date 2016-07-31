
//window.onresize = doLayout;

var def_pool = 'http://pool2.dcoin.club';

onload = function() {
     chrome.storage.local.get('pool',function( obj ) {
     	if ( obj['pool'] ) {
     		def_pool = obj['pool'];
     		go_to_pool();
     	}
     });
}

/*chrome.storage.local.set({'pool': def_pool}, function() {
          chrome.storage.local.get('pool',function( obj ) {
   			console.log( obj);
        });
});*/

function go_to_pool() {
	chrome.storage.local.set({'pool': def_pool});
	$('body').html( '<webview id="webview" src="' + def_pool + '"></webview>' );
	window.onresize = doLayout;
	doLayout();
}

$('#change_pkey_key_file_name').on('click',function(e){
	document.getElementById('change_pkey_upload_hidden').click();
});

$('#key_btn').on('click',function(e){
	document.getElementById('change_pkey_upload_hidden').click();
});

$('#send_btn').on('click',function(e){
	send_key();
});

$('#next_btn').on('click',function(e) {
	go_to_pool();
});

$('#change_pkey_key_selector a').on('click',function(e){
	change_pkey_show_text_key();
});

function doLayout() {
  var webview = document.querySelector('webview');
  var windowWidth = document.documentElement.clientWidth;
  var windowHeight = document.documentElement.clientHeight;
  var webviewWidth = windowWidth;
  var webviewHeight = windowHeight;

  webview.style.width = webviewWidth + 'px';
  webview.style.height = webviewHeight + 'px';
}

		(function ($) {
		$.fn.spin = function (opts, color) {
			var presets = {
				"tiny": {
					lines: 8,
					length: 2,
					width: 2,
					radius: 3
				},
				"small": {
					lines: 8,
					length: 4,
					width: 3,
					radius: 5
				},
				"large": {
					lines: 10,
					length: 8,
					width: 4,
					radius: 8
				}
			};
			if (Spinner) {
				return this.each(function () {
					var $this = $(this),
							data = $this.data();

					if (data.spinner) {
						data.spinner.stop();
						delete data.spinner;
					}
					if (opts !== false) {
						if (typeof opts === "string") {
							if (opts in presets) {
								opts = presets[opts];
							} else {
								opts = {};
							}
							if (color) {
								opts.color = color;
							}
						}
						data.spinner = new Spinner($.extend({
							color: $this.css('color')
						}, opts)).spin(this);
					}
				});
			} else {
				throw "Spinner class not available.";
			}
		};
	})(jQuery);

			function send_key() {
				var e_n_sign = get_e_n_sign($("#change_pkey_private_key").val(), $("#change_pkey_password").val(), '', 'change_pkey_alert');
				if ( e_n_sign['modulus'] != '' || e_n_sign['exp']!='' ) {
					var public_key = make_public_key(e_n_sign['modulus'], e_n_sign['exp']);
					$.post( 'http://getpool.dcoin.club/', {
						'public_key' : public_key
					}, function(data) {
						answer = JSON.parse( data );
						if (answer['pool'] && answer['pool'].length > 7)
							def_pool = answer['pool'];
						go_to_pool()
					});
				}
			}
		
			function change_pkey_show_text_key () {
				$("#change_pkey_private_key").css("display", "block");
				$("#change_pkey_key_div").css("display", "none");
				$("#change_pkey_key_selector").html('<a href="#" id="from_file" onclickx="change_pkey_show_file_key();return false;">From the file</a>');
				$('#from_file').on('click',function(e){
					change_pkey_show_file_key();
				});

			}

			function change_pkey_show_file_key () {
				$("#change_pkey_private_key").css("display", "none");
				$("#change_pkey_key_div").css("display", "block");
				$("#change_pkey_key_selector").html('<a href="#" id="from_text" onclickx="change_pkey_show_text_key();return false;">Text</a>');
				$('#from_text').on('click',function(e){
					change_pkey_show_text_key();
				});
			}
	
			function change_handleFileSelect(f) {
				$('#change_pkey_key_file_name').html(f.name);
				var reader = new FileReader();
				if (f.type.substr(0,5) =='image') {
					reader.onload = (function(theFile) {
						return function(e) {
							img2key(e.target.result, 'change_pkey_private_key');
						};
					})(f);
					reader.readAsDataURL(f);
				}
				else {
					reader.onload = (function(theFile) {
						return function(e) {
							console.log(e.target.result);
							$('#change_pkey_private_key').val(e.target.result);
						};
					})(f);
					reader.readAsText(f);
				}
			}

	
			$( document ).ready(function() {
				if (window.FileReader === undefined) {
					$("#change_pkey_private_key").css("display", "block");
					$("#change_pkey_key_file").css("display", "none");
					$("#change_pkey_key_selector").css("display", "none");
				}
//				$("#tx_history").css("display", "block");
//				show_steps('simple_protection_mode');

				document.getElementById('change_pkey_upload_hidden').addEventListener('change', change_handleFileSelect2, false);
//				check_key_and_show_modal();

			});

			function change_handleFileSelect2(evt) {
				$('#change_pkey_key_file_name').html(this.value);
				var f = evt.target.files[0];
				change_handleFileSelect(f);
			}

			$('#change_pkey_key_div').on(
					'dragover',
					function(e) {
						e.preventDefault();
						e.stopPropagation();
					}
			)
			$('#change_pkey_key_div').on(
					'dragenter',
					function(e) {
						e.preventDefault();
						e.stopPropagation();
					}
			)
			$('#change_pkey_key_div').on(
					'drop',
					function(e){
						if(e.originalEvent.dataTransfer){
							if(e.originalEvent.dataTransfer.files.length) {
								e.preventDefault();
								e.stopPropagation();
								change_handleFileSelect(e.originalEvent.dataTransfer.files[0]);
							}
						}
					}
			);
