var webpage = "";
function supportstorage() {
	return typeof window.localStorage === 'object';
	/*if (typeof window.localStorage=='object') {
		return true;
	} else {
		return false;
	}*/
}

function handleSaveLayout() {
	var e = $(".demo").html();
	if (!stopsave && e != window.demoHtml) {
		stopsave++;
		window.demoHtml = e;
		saveLayout();
		stopsave--;
	}
}

var layouthistory;
function saveLayout(){
	var data = layouthistory;
	if (!data) {
		data={};
		data.count = 0;
		data.list = [];
	}
	if (data.list.length>data.count) {
		for (i=data.count;i<data.list.length;i++)
			data.list[i]=null;
	}
	
	data.list[data.count] = window.demoHtml;
	data.count++;
	if (supportstorage()) {
		localStorage.setItem("eGaaSEditor",JSON.stringify(data));
	}
	layouthistory = data;
}

function undoLayout() {
	var data = layouthistory;
	//console.log(data);
	if (data) {
		if (data.count<2) return false;
		window.demoHtml = data.list[data.count-2];
		data.count--;
		$('.demo').html(window.demoHtml);
		if (supportstorage()) {
			localStorage.setItem("eGaaSEditor",JSON.stringify(data));
		}
		return true;
	}
	return false;
}

function redoLayout() {
	var data = layouthistory;
	if (data) {
		if (data.list[data.count]) {
			window.demoHtml = data.list[data.count];
			data.count++;
			$('.demo').html(window.demoHtml);
			if (supportstorage()) {
				localStorage.setItem("eGaaSEditor",JSON.stringify(data));
			}
			return true;
		}
	}
	return false;
}

function handleJsIds() {
	handleModalIds();
	handleAccordionIds();
	handleCarouselIds();
	handleTabsIds();
}
function handleAccordionIds() {
	var e = $(".demo #myAccordion");
	var t = randomNumber();
	var n = "accordion-" + t;
	var r;
	e.attr("id", n);
	e.find(".accordion-group").each(function(e, t) {
		r = "accordion-element-" + randomNumber();
		$(t).find(".accordion-toggle").each(function(e, t) {
			$(t).attr("data-parent", "#" + n);
			$(t).attr("href", "#" + r);
		});
		$(t).find(".accordion-body").each(function(e, t) {
			$(t).attr("id", r);
		});
	});
}
function handleCarouselIds() {
	var e = $(".demo #myCarousel");
	var t = randomNumber();
	var n = "carousel-" + t;
	e.attr("id", n);
	e.find(".carousel-indicators li").each(function(e, t) {
		$(t).attr("data-target", "#" + n);
	});
	e.find(".left").attr("href", "#" + n);
	e.find(".right").attr("href", "#" + n);
}
function handleModalIds() {
	var e = $(".demo #myModalLink");
	var t = randomNumber();
	var n = "modal-container-" + t;
	var r = "modal-" + t;
	e.attr("id", r);
	e.attr("href", "#" + n);
	e.next().attr("id", n);
}
function handleTabsIds() {
	var e = $(".demo #myTabs");
	var t = randomNumber();
	var n = "tabs-" + t;
	e.attr("id", n);
	e.find(".tab-pane").each(function(e, t) {
		var n = $(t).attr("id");
		var r = "panel-" + randomNumber();
		$(t).attr("id", r);
		$(t).parent().parent().find("a[href=#" + n + "]").attr("href", "#" + r);
	});
}
function randomNumber() {
	return randomFromInterval(1, 1e6);
}
function randomFromInterval(e, t) {
	return Math.floor(Math.random() * (t - e + 1) + e);
}
/*function gridSystemGenerator() {
	$(".lyrow .preview input").bind("keyup", function() {
		var e = 0;
		var t = "";
		var n = $(this).val().split(" ", 12);
		$.each(n, function(n, r) {
			e = e + parseInt(r);
			t += '<div class="span' + r + ' column"></div>';
		});
		if (e == 12) {
			$(this).parent().next().children().html(t);
			$(this).parent().prev().show();
		} else {
			$(this).parent().prev().hide();
		}
	});
}*/
function configurationElm(e, t) {
	$(".demo").delegate(".configuration > a", "click", function(e) {
		e.preventDefault();
		var t = $(this).parent().next().next().children();
		$(this).toggleClass("active");
		t.toggleClass($(this).attr("rel"));
	});
	$(".demo").delegate(".configuration .dropdown-menu a", "click", function(e) {
		e.preventDefault();
		
		var t = $(this).parent().parent();
		var n = t.parent().parent().next().next().children();
		
		if ($(this).hasClass("remove") || $(this).hasClass("editor")) {
			$(this).parent().removeClass("active");
		} else {
			t.find("li").removeClass("active");
			$(this).parent().addClass("active");
			var r = "";
			t.find("a").each(function() {
				r += $(this).attr("rel") + " ";
			});
			t.parent().removeClass("open");
			n.removeClass(r);
			n.addClass($(this).attr("rel"));
		}
	});
}
function removeElm() {
	$(".demo").delegate(".remove", "click", function(e) {
		e.preventDefault();
		$(this).closest(".lyrow").remove();
		/*if (!$(".demo .lyrow").length > 0) {
			clearDemo();
		}*/
	});
}
function clearDemo() {
	$(".demo").empty();
	layouthistory = null;
	if (supportstorage())
		localStorage.removeItem("eGaaSEditor");
}
function removeMenuClasses() {
	$("#menu-layoutit li button").removeClass("active");
}
function changeStructure(e, t) {
	$("#download-layout ." + e).removeClass(e).addClass(t);
}
function cleanHtml(e) {
	$(e).parent().append($(e).children().html());
}

var currentDocument = null;
var timerSave = 1000;
var stopsave = 0;
var startdrag = 0;
var demoHtml = $(".demo").html();
var currenteditor = null;
var connectedStatus = false;
// $(window).resize(function() {
// 	$("body").css("min-height", $(window).height() - 90);
// 	$(".demo").css("min-height", $(window).height() - 160)
// });

function restoreData(){
	if (supportstorage()) {
		layouthistory = JSON.parse(localStorage.getItem("eGaaSEditor"));
		if (!layouthistory) {
			return false;
		}
		//console.log(layouthistory);
		window.demoHtml = layouthistory.list[layouthistory.count-1];
		if (window.demoHtml) {
			//console.log($(".demo"));
			$(".demo").html(window.demoHtml);
			$(".demo").find(".ui-resizable-handle.ui-resizable-e").remove();
		}
	}
}

function initContainer(){
	$(".demo, .demo .column").sortable({
		//connectWith: ".column",
		opacity: .35,
		handle: ".drag",
		placeholder: "portlet-placeholder ui-corner-all",
		start: function(event, ui) {
			var col = $(ui.item).attr("class").split(" ");
			for (var i = 0; i < col.length; i++) {
				var temp = col[i].split("-");
				if (temp.length === 3 && temp[0] === "col") {
					$(ui.placeholder).addClass("col-" + temp[1] + "-" + temp[2]);
				}
			}
			
			if (!startdrag) stopsave++;
			startdrag = 1;
		},
		stop: function(event, ui) {
			if(stopsave>0) stopsave--;
			startdrag = 0;
		}
	});
	
	$(".demo .lyrow").resizable({
		handles: "e",
		containment: "parent",
		resize: function(event, ui) {
			//console.log(ui);
			var w = $(ui.element).parent().width();
			var d = Math.ceil((ui.size.width * 12) / w);
			if (d < 1) {
				d = 1;
			} else if (d > 12) {
				d = 12;
			}
			var col = $(ui.element).attr("class").split(" ");
			for (var i = 0; i < col.length; i++) {
				var temp = col[i].split("-");
				if (temp.length === 3 && temp[0] === "col") {
					$(ui.element).removeClass("col-" + temp[1] + "-" + temp[2]);
					$(ui.element).addClass("col-" + temp[1] + "-" + d);
				}
			}
		}
	});
	
	Connect();
	$("#connect").on('click', function(){
		Connect();
	});
	
	configurationElm();
}

function Connect(){
	 "use strict";
	
	if ($("#connect").prop("checked") === true) {
		//$("#connect").parent().parent().parent().parent().parent().find(".view .column").addClass("connected");
		$(".demo, .demo .column").sortable("option", "connectWith", ".column");
	} else {
		//$("#connect").parent().parent().parent().parent().parent().find(".view .column").removeClass("connected");
		$(".demo, .demo .column").sortable("option", "connectWith", "");
	}
}

function initGenerator(){
  "use strict";
	
	restoreData();
	
	if (CKEDITOR.instances.contenteditor) {
		CKEDITOR.instances.contenteditor.destroy();
	}
	
	CKEDITOR.disableAutoInline = true;
	var contenthandle = CKEDITOR.replace( 'contenteditor' ,{
		language: 'en',
		contentsCss: ['static/css/style.css'],
		allowedContent: true
	});
	
	$("#getColumnViewport, #getColumnGrid").on('change', function(){
		$(this).parent().parent().parent().attr("class", "col-" + $("#getColumnViewport").val() + "-" + $("#getColumnGrid").val() + " lyrow");
	});
	
	$(".sidebar-nav .lyrow").draggable({
		connectToSortable: ".demo",
		helper: "clone",
		handle: ".drag",
		placeholder: "portlet-placeholder ui-corner-all",
		start: function(event, ui) {
			if (!startdrag) stopsave++;
			startdrag = 1;
		},
		drag: function(event, ui) {
			ui.helper.width(400);
		},
		stop: function(event, ui) {
			$(".demo .lyrow .preview").remove();
			$(".demo .lyrow .drag").removeClass("column");

			$(".demo .column").sortable({
				opacity: .35,
				handle: ".drag",
				placeholder: "portlet-placeholder ui-corner-all",
				start: function(event, ui) {
					var col = $(ui.item).attr("class").split(" ");
					
					for (var i = 0; i < col.length; i++) {
						var temp = col[i].split("-");
						if (temp.length === 3 && temp[0] === "col") {
							$(ui.placeholder).addClass("col-" + temp[1] + "-" + temp[2]);
						}
					}
					
					if (!startdrag) stopsave++;
					startdrag = 1;
				},
				stop: function(event, ui) {
					if(stopsave>0) stopsave--;
					startdrag = 0;
				}
			});
			$(".demo .lyrow").resizable({
				handles: "e",
				containment: "parent",
				resize: function(event, ui) {
					//console.log(ui);
					var w = $(ui.element).parent().width();
					var d = Math.ceil((ui.size.width * 12) / w);
					if (d < 1) {
						d = 1;
					} else if (d > 12) {
						d = 12;
					}
					var col = $(ui.element).attr("class").split(" ");
					for (var i = 0; i < col.length; i++) {
						var temp = col[i].split("-");
						if (temp.length === 3 && temp[0] === "col") {
							$(ui.element).removeClass("col-" + temp[1] + "-" + temp[2]);
							$(ui.element).addClass("col-" + temp[1] + "-" + d);
						}
					}
				}
			});
			if(stopsave>0) stopsave--;
			startdrag = 0;
		}
	});
	$(".sidebar-nav .box").draggable({
		connectToSortable: ".column",
		helper: "clone",
		handle: ".drag",
		placeholder: "portlet-placeholder ui-corner-all",
		start: function(e,t) {
			if (!startdrag) stopsave++;
			startdrag = 1;
		},
		drag: function(e, t) {
			t.helper.width(400);
		},
		stop: function() {
			handleJsIds();
			if(stopsave>0) stopsave--;
			startdrag = 0;
		}
	});
	
	$(".selectbox").select2({
		dropdownParent: $("#dl_content"),
		minimumResultsForSearch: Infinity,
		theme: 'bootstrap'
	});
	
	initContainer();
	
	$("#editorModal").on('show.bs.modal', function (e) {
		currenteditor = $(e.relatedTarget).closest(".lyrow").find('.view .column');
		var eText = currenteditor.html();
		contenthandle.setData(eText);
	})
	$("#savecontent").on('click', function() {
		currenteditor.html(contenthandle.getData());
		$("#editorModal").modal('hide');
	});
	$("[data-target=#downloadModal]").click(function(e) {
		e.preventDefault();
		downloadLayoutSrc();
	});
	$("[data-target=#shareModal]").click(function(e) {
		e.preventDefault();
		handleSaveLayout();
	});
	$("#download").click(function() {
		downloadLayout();
		return false;
	});
	$("#downloadhtml").click(function() {
		downloadHtmlLayout();
		return false;
	});
	$("#edit").click(function() {
		$("body").removeClass("devpreview sourcepreview");
		$("body").addClass("edit");
		removeMenuClasses();
		$(this).addClass("active");
		return false;
	});
	$("#clear").click(function(e) {
		e.preventDefault();
		clearDemo();
	});
	$("#devpreview").click(function() {
		$("body").removeClass("edit sourcepreview");
		$("body").addClass("devpreview");
		removeMenuClasses();
		$(this).addClass("active");
		return false;
	});
	$("#sourcepreview").click(function() {
		$("body").removeClass("edit");
		$("body").addClass("devpreview sourcepreview");
		removeMenuClasses();
		$(this).addClass("active");
		return false;
	});
	$("#fluidPage").click(function(e) {
		e.preventDefault();
		changeStructure("container", "container-fluid");
		$("#fixedPage").removeClass("active");
		$(this).addClass("active");
		downloadLayoutSrc();
	});
	$("#fixedPage").click(function(e) {
		e.preventDefault();
		changeStructure("container-fluid", "container");
		$("#fluidPage").removeClass("active");
		$(this).addClass("active");
		downloadLayoutSrc();
	});
	$(".nav-header").click(function() {
		$(".sidebar-nav .boxes, .sidebar-nav .rows").hide();
		$(this).next().slideDown();
	});
	$('#undo').click(function(){
		stopsave++;
		if (undoLayout()) initContainer();
		stopsave--;
	});
	$('#redo').click(function(){
		stopsave++;
		if (redoLayout()) initContainer();
		stopsave--;
	});
	removeElm();
	//gridSystemGenerator();
	var handleSaveLayoutInterval = setInterval(function() {
		if ($("#eGaaSEditor").length) {
			handleSaveLayout();
		} else {
			clearInterval(handleSaveLayoutInterval);
		}
	}, timerSave);
}

initGenerator();

function saveHtml()
			{
                        var cpath = window.location.href;
                        cpath = cpath.substring(0, cpath.lastIndexOf("/"));
			webpage = '<html>\n<head>\n<script type="text/javascript" src="'+cpath+'/js/jquery-2.0.0.min.js"></script>\n<script type="text/javascript" src="'+cpath+'/js/jquery-ui.js"></script>\n<link href="'+cpath+'/css/bootstrap-combined.min.css" rel="stylesheet" media="screen">\n<script type="text/javascript" src="'+cpath+'/js/bootstrap.min.js"></script>\n</head>\n<body>\n'+ webpage +'\n</body>\n</html>'
			/* FM aka Vegetam Added the function that save the file in the directory Downloads. Work only to Chrome Firefox And IE*/
			if (navigator.appName =="Microsoft Internet Explorer" && window.ActiveXObject)
			{
			var locationFile = location.href.toString();
			var dlg = false;
			with(document){
			ir=createElement('iframe');
			ir.id='ifr';
			ir.location='about.blank';
			ir.style.display='none';
			body.appendChild(ir);
			with(getElementById('ifr').contentWindow.document){
			open("text/html", "replace");
			charset = "utf-8";
			write(webpage);
			close();
			document.charset = "utf-8";
			dlg = execCommand('SaveAs', false, locationFile+"webpage.html");
			}
    return dlg;
			}
			}
			else{
			webpage = webpage;
			var blob = new Blob([webpage], {type: "text/html;charset=utf-8"});
			saveAs(blob, "webpage.html");
		}
		}
