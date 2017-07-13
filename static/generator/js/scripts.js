var eGaaSID = false;
var webpage = "";

function supportstorage() {
	return typeof window.localStorage === 'object';
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

function configurationFocus(el) {
	"use strict";
	
	var _this = el;
	var id = _this.attr("egaas-id");
	var elem = $('[egaas-for="' + id + '"]');
	
	/*if (_this.prop("tagName") && _this.prop("tagName").toLowerCase() === "button") {
		jQuery('<div/>').attr("egaas-id", id).insertBefore(_this);
		_this.removeAttr("egaas-id");
		_this.appendTo('[egaas-id="' + id + '"]');
	} else {
		elem.prependTo(_this);
	}*/
	
	elem.prependTo(_this);
	
	if (elem.is(":hidden")) {
		if (_this.hasClass("table-responsive")) {
			elem.show().css({"left": "0px", "top": "0px", "width": "100%"});
		} else {
			elem.show().css({"left": "0px", "top": -elem.height() + "px", "width": "100%"});
		}
	}
}
function configurationBlur(el) {
	"use strict";
	
	var _this = el;
	var id = _this.attr("egaas-id");
	var elem = $('[egaas-for="' + id + '"]');

	//elem.insertBefore(_this).hide();
	elem.hide();
}
function configurationElm() {
	"use strict";
	
	/*$(".demo").delegate(".configuration > a", "click", function(e) {
		e.preventDefault();
		var t = $(this).parent().next().next().children();
		$(this).toggleClass("active");
		t.toggleClass($(this).attr("rel"));
	});*/
	
	$(".demo").delegate(".configuration .dropdown-menu a", "click", function(e) {
		e.preventDefault();
		
		var elem;
		var style = "";
		var cont = $(this).parent().parent();
		var id = $(this).closest(".configuration").attr("egaas-for");
		
		if (id) {
			elem = $('[egaas-id="' + id + '"]').children(":not(.configuration)");
		} else {
			elem = cont.closest(".configuration").parent().find(".view .column").children();
		}
		
		if ($(this).hasClass("remove") || $(this).hasClass("editor") || $(this).hasClass("submenu") || $(this).hasClass("toggle")) {
			if ($(this).hasClass("toggle")) {
				$(this).parent().toggleClass("active");
				if (elem.hasClass("table-responsive")) {
					elem = elem.children();
				}
				
				if (elem.prop("tagName").toLowerCase() === "input" && $(this).attr("rel") === "disabled") {
					if (elem.prop("disabled")) {
						elem.prop("disabled", false);
					} else {
						elem.prop("disabled", true);
					}
				} else {
					elem.toggleClass($(this).attr("rel"));
				}
				
				return false;
			} else {
				if ($(this).hasClass("submenu")) {
					return false;
				}
				$(this).parent().removeClass("active");
			}
		} else {
			cont.find("> li").removeClass("active");
			$(this).parent().addClass("active");
			
			cont.find("a").each(function() {
				style += $(this).attr("rel") + " ";
			});
			
			cont.parent().removeClass("open");
			elem.removeClass(style);
			elem.addClass($(this).attr("rel"));
			
			return false;
		}
	});
	
	$(".demo").delegate(".configuration .dropdown-menu .styles", "mouseenter", function() {
		var cont = $(this);
		var pop = cont.find(".dropdown-menu:first");
		
		if (pop.is(":hidden")) {
			pop.show();
		}
	});
	
	$(".demo").delegate(".configuration .dropdown-menu .styles", "mouseleave", function() {
		var cont = $(this);
		var pop = cont.find(".dropdown-menu:first");

		pop.hide();
	});
	
	$(".demo").delegate("[egaas-id]", "mouseenter", function() {
		configurationFocus($(this));
	});
	
	$(".demo").delegate("[egaas-id]", "mouseleave", function() {
		configurationBlur($(this));
	});
}
function removeElm() {
	"use strict";
	
	$(".demo").delegate(".remove", "click", function(e) {
		e.preventDefault();
		$(this).closest(".lyrow").remove();
		/*if (!$(".demo .lyrow").length > 0) {
			clearDemo();
		}*/
	});
}
function clearDemo() {
	"use strict";
	
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
	"use strict";
	
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
			$(".demo").find(".open").removeClass("open");
		}
	}
}

function initContainer(){
	"use strict";
	
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
	$(".demo .panel-body").sortable({
		connectWith: ".panel-body",
		opacity: .35,
		handle: ".drag",
		placeholder: "portlet-placeholder ui-corner-all",
		start: function(event, ui) {
			eGaaSID = $(ui.helper).attr("egaas-for");
			
			if (!startdrag) stopsave++;
			startdrag = 1;
		},
		stop: function(event, ui) {
			var el = $('[egaas-id="' + eGaaSID + '"]');
			el.insertAfter($(ui.item));
			configurationFocus(el);
			
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
	
	var innerContainer;
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
			innerContainer = $(".portlet-placeholder").parent();
			ui.helper.width(400);
		},
		stop: function(event, ui) {
			$(".demo .lyrow .preview").remove();
			$(".demo .lyrow .drag").removeClass("column");
			//console.log(ui.helper);
			
			if ($("#connect").prop("checked") === true) {
				var grid = false;
				var el = $(ui.helper).find(".column").html();
				var settings = $(ui.helper).find(".settings ul").html();
				var col = ui.helper[0].classList;
				
				for (var i = 0; i < col.length; i++) {
					var temp = col[i].split("-");
					if (temp.length === 3 && temp[0] === "col") {
						grid = true;
					}
				}
				
				if (!grid) {
					innerContainer.find(".lyrow").remove();
					innerContainer.closest(".lyrow").find(".settings ul").prepend(settings);
					innerContainer.html(el);
				}
			}
			
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
			$(".demo .panel-body").sortable({
				connectWith: ".panel-body",
				opacity: .35,
				handle: ".drag",
				placeholder: "portlet-placeholder ui-corner-all",
				start: function(event, ui) {
										
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
		connectToSortable: ".panel-body",
		helper: "clone",
		handle: ".drag",
		placeholder: "portlet-placeholder ui-corner-all",
		start: function(event, ui) {
			if (!startdrag) stopsave++;
			startdrag = 1;
		},
		drag: function(event, ui) {
			$(".portlet-placeholder").addClass($(ui.helper).attr("egaas-class"));
			innerContainer = $(".portlet-placeholder").parent();
			ui.helper.width(400);
		},
		stop: function(event, ui) {
			var id = new Date().getTime();
			var settings = $(ui.helper).find(".settings").parent();
			var el = $(ui.helper).find(".column").html();
			
			//$(".demo .box .preview").remove();
			settings.find(".drag").html("");
			$(settings).attr("egaas-for", "element-" + id).insertBefore(innerContainer.find(".box"));
			//$($(el)).attr("egaas-id", "element-" + id).insertBefore(innerContainer.find(".box"));
			jQuery('<div/>').attr("egaas-id", "element-" + id).insertBefore(innerContainer.find(".box"));
			$(el).appendTo('[egaas-id="element-' + id + '"]');
			innerContainer.find(".box").remove();
			
			$(".demo .panel-body").sortable({
				connectWith: ".panel-body",
				opacity: .35,
				handle: ".drag",
				placeholder: "portlet-placeholder ui-corner-all",
				start: function(event, ui) {
					eGaaSID = $(ui.helper).attr("egaas-for");

					if (!startdrag) stopsave++;
					startdrag = 1;
				},
				stop: function(event, ui) {
					var el = $('[egaas-id="' + eGaaSID + '"]');
					el.insertAfter($(ui.item));
					configurationFocus(el);

					if(stopsave>0) stopsave--;
					startdrag = 0;
				}
			});
			
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

function resizeCanvas(size) {
	var containerID = document.getElementsByClassName("changeDimension");
	if (size == "lg") {
		$(containerID).attr('id', "LG");
	}
	if (size == "md") {
		$(containerID).attr('id', "MD");
	}
	if (size == "sm") {
		$(containerID).attr('id', "SM");
	}
	if (size == "xs") {
		$(containerID).attr('id', "XS");
	}
}

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
