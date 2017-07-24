var eGaaSID = false;
var layouthistory;
var timerSave = 1000;
var stopsave = 0;
var startdrag = 0;
var demoHtml = $(".demo").html();
var currenteditor = null;
var configPanel = null;
var CKEDITOR;

function templateDIV(viewport, grid, content, config) {
	"use strict";
	
	return '<div class="lyrow col-' + viewport + '-' + grid + '">' +
				config +
				'<div class="view">' +
					'<div class="column">' +
						content +
					'</div>' +
				'</div>' +
			'</div>';
}

function supportstorage() {
	"use strict";
	
	return typeof window.localStorage === 'object';
}

function handleSaveLayout() {
	"use strict";
	
	var e = $(".demo").html();
	if (!stopsave && e !== window.demoHtml) {
		stopsave++;
		window.demoHtml = e;
		saveLayout();
		stopsave--;
	}
}

function saveLayout(){
	"use strict";
	
	var data = layouthistory;
	if (!data) {
		data={};
		data.count = 0;
		data.list = [];
	}
	if (data.list.length>data.count) {
		for (var i = data.count; i < data.list.length; i++) {
			data.list[i] = null;
		}
	}
	
	data.list[data.count] = window.demoHtml;
	data.count++;
	if (supportstorage()) {
		localStorage.setItem("eGaaSEditor",JSON.stringify(data));
	}
	layouthistory = data;
}

function undoLayout() {
	"use strict";
	
	var data = layouthistory;
	//console.log(data);
	if (data) {
		if (data.count < 2) {
			return false;
		}
		
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
	"use strict";
	
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
	elem.show();
	
	/*if (elem.is(":hidden")) {
		if (_this.hasClass("table-responsive")) {
			elem.show().css({"left": "0px", "top": "0px", "width": "100%"});
		} else {
			elem.show().css({"left": "0px", "top": -elem.height() + "px", "width": "100%"});
		}
	}*/
}

function configurationBlur(el) {
	"use strict";
	
	var _this = el;
	var id = _this.attr("egaas-id");
	var elem = $('[egaas-for="' + id + '"]');

	//elem.insertBefore(_this).hide();
	elem.hide();
}

function configElements() {
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
				
				if ($(this).attr("rel") === "disabled") {
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
			cont.find("> li a:not(.toggle)").parent().removeClass("active");
			$(this).parent().addClass("active");
			
			cont.find("a:not(.toggle)").each(function() {
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
		
		if ($(this).closest("[egaas-id]").length > 0) {
			$(this).closest("[egaas-id]").remove();
		} else {
			$(this).closest(".lyrow").remove();
		}
		
		/*if (!$(".demo .lyrow").length > 0) {
			clearDemo();
		}*/
	});
}
function clearDemo() {
	"use strict";
	
	$(".demo").empty();
	layouthistory = null;
	
	if (supportstorage()) {
		localStorage.removeItem("eGaaSEditor");
	}
}

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
		opacity: 0.35,
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
			
			if (!startdrag) {
				stopsave++;
			}
			
			startdrag = 1;
		},
		stop: function() {
			if(stopsave > 0) {
				stopsave--;
			}
			
			startdrag = 0;
		}
	});
	$(".demo .panel-body").sortable({
		connectWith: ".panel-body",
		opacity: 0.35,
		handle: ".drag",
		placeholder: "portlet-placeholder ui-corner-all",
		start: function(event, ui) {
			eGaaSID = $(ui.helper).attr("egaas-for");
			
			if (!startdrag) {
				stopsave++;
			}
			
			startdrag = 1;
		},
		stop: function(event, ui) {
			var el = $('[egaas-id="' + eGaaSID + '"]');
			el.insertAfter($(ui.item));
			configurationFocus(el);
			
			if(stopsave > 0) {
				stopsave--;
			}
			
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
	
	$("#connect").on('click', function(){
		Connect();
	});
	
	Connect();
	configElements();
}

function Connect(){
	 "use strict";
	
	if ($("#connect").prop("checked") === true) {
		$(".demo, .demo .column").sortable("option", "connectWith", ".column");
	} else {
		$(".demo, .demo .column").sortable("option", "connectWith", "");
	}
}

function initGenerator(){
  "use strict";
	
	var innerContainer;
	
	restoreData();
	
	if (CKEDITOR.instances.contenteditor) {
		CKEDITOR.instances.contenteditor.destroy();
	}
	
	CKEDITOR.disableAutoInline = true;
	var contenthandle = CKEDITOR.replace( 'contenteditor' ,{
		language: 'en',
		contentsCss: ['static/css/style.css']
	});
	
	$(".selectbox").select2({
		dropdownParent: $("#dl_content"),
		minimumResultsForSearch: Infinity,
		theme: 'bootstrap'
	});
	
	$("#getColumnViewport, #getColumnGrid").on('change', function(){
		$(this).parent().parent().parent().attr("class", "col-" + $("#getColumnViewport").val() + "-" + $("#getColumnGrid").val() + " lyrow");
	});
	
	$(".sidebar-nav .lyrow").draggable({
		connectToSortable: ".demo",
		helper: "clone",
		handle: ".drag",
		placeholder: "portlet-placeholder ui-corner-all",
		start: function() {
			if (!startdrag) {
				stopsave++;
			}
			
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
				opacity: 0.35,
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
					
					if (!startdrag) {
						stopsave++;
					}
					
					startdrag = 1;
				},
				stop: function() {
					if(stopsave > 0) {
						stopsave--;
					}
					
					startdrag = 0;
				}
			});
			$(".demo .panel-body").sortable({
				connectWith: ".panel-body",
				opacity: 0.35,
				handle: ".drag",
				placeholder: "portlet-placeholder ui-corner-all",
				start: function() {
					if (!startdrag) {
						stopsave++;
					}
					
					startdrag = 1;
				},
				stop: function() {
					if(stopsave > 0) {
						stopsave--;
					}
					
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
			
			if(stopsave > 0) {
				stopsave--;
			}
			
			startdrag = 0;
		}
	});
	$(".sidebar-nav .box").draggable({
		connectToSortable: ".panel-body",
		helper: "clone",
		handle: ".drag",
		placeholder: "portlet-placeholder ui-corner-all",
		start: function(event, ui) {
			if ($(ui.helper).attr("egaas-class") === "table") {
				if ($("#getTableType").val() === '0') {
					$(ui.helper).find(".table-responsive").append(
						'<table class="table" data-role="table">' +
							'<thead>' +
								'<tr>' +
									'<th>#</th>' +
									'<th>First Name</th>' +
									'<th>Last Name</th>' +
									'<th>Username</th>' +
								'</tr>' +
							'</thead>' +
							'<tbody>' +
								'<tr>' +
									'<td>1</td>' +
									'<td>Mark</td>' +
									'<td>Otto</td>' +
									'<td>@mdo</td>' +
								'</tr>' +
								'<tr>' +
									'<td>2</td>' +
									'<td>Jacob</td>' +
									'<td>Thornton</td>' +
									'<td>@fat</td>' +
								'</tr>' +
								'<tr>' +
									'<td>3</td>' +
									'<td>Larry</td>' +
									'<td>the Bird</td>' +
									'<td>@twitter</td>' +
								'</tr>' +
							'</tbody>' +
						'</table>'
					);
				} else {
					var name = $("#getTableDataName").val();
					var order = $("#getTableDataOrder").val() ? 'Order: ' + $("#getTableDataOrder").val() + ' ' : "";
					var where = $("#getTableDataWhere").val() ? 'Where: ' + $("#getTableDataWhere").val() + ' ' : "";
					var columns = "";
					
					$("#getTableData").find(".Columns").each(function(index, element){
						var col = '[';
						
						$(element).find(".form-control").each(function(index, elem){
							if ($(elem).val()) {
								//col.push($(elem).val());
								if (index === 0) {
									col = $(elem).val();
								/*} else if (index === $(element).find(".form-control").length - 1) {
									col = col + ', ' + $(elem).val() + ']';*/
								} else {
									col = col + ', ' + $(elem).val();
								}
							}
						});
						//columns.push(col);
						if (index === 0) {
							columns = col + ']';
						} else {
							columns = columns + ', ' + col + ']';
						}
					});
					
					$(ui.helper).find(".table-responsive").append(
						'Table {' +
							'Class: data-role="table"' +
							'Table: ' + name + ' ' + 
							order +
							where +
							'Columns: [' +
								columns +
							']' +
						'}'
					);
				}
			}
			if (!startdrag) {
				stopsave++;
			}
			
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
				opacity: 0.35,
				handle: ".drag",
				placeholder: "portlet-placeholder ui-corner-all",
				start: function(event, ui) {
					eGaaSID = $(ui.helper).attr("egaas-for");

					if (!startdrag) {
						stopsave++;
					}
					
					startdrag = 1;
				},
				stop: function(event, ui) {
					var el = $('[egaas-id="' + eGaaSID + '"]');
					el.insertAfter($(ui.item));
					configurationFocus(el);

					if(stopsave > 0) {
						stopsave--;
					}
					
					startdrag = 0;
				}
			});
			
			if(stopsave > 0) {
				stopsave--;
			}
			
			startdrag = 0;
		}
	});
		
	$("#editorModal").on('show.bs.modal', function (e) {
		var editText;
		
		if ($(e.relatedTarget).attr("id") !== "UploadHTML") {
			if ($(e.relatedTarget).closest("[egaas-id]").length > 0) {
				currenteditor = $(e.relatedTarget).closest("[egaas-id]");
				configPanel = currenteditor.find(".configuration")[0].outerHTML;
			} else {
				if ($(e.relatedTarget).closest(".lyrow").find('.view .column .panel').length > 0) {
					currenteditor = $(e.relatedTarget).closest(".lyrow").find('.view .column .panel .panel-body');
				} else {
					currenteditor = $(e.relatedTarget).closest(".lyrow").find('.view .column');
				}
			}
			editText = currenteditor.html().replace(configPanel, "");
		} else {
			currenteditor = $(".demo");
			editText = "";
		}
		
		contenthandle.setData(editText);
	});
	
	$("#savecontent").on('click', function() {
		if (!currenteditor.hasClass("row")) {
			currenteditor.html(configPanel + contenthandle.getData());
		} else {
			currenteditor.html(contenthandle.getData());
			console.log(currenteditor.find("*"));
			var tags = currenteditor.find("*");
			var deep = 0;
			
			for (var k = 0; k < tags.length; k++) {
				var tagTemp = $(tags[k]);
				var deepTemp = tagTemp.parentsUntil(currenteditor).length;
				
				if (deepTemp > deep) {
					deep = deepTemp;
				}
			}
			
			for (var d = deep; d >= 0; d--) {
				for (var i = 0; i < tags.length; i++) {
					var tag = $(tags[i]);
					var tagName = tags[i].tagName.toLowerCase();
					
					if (tag.parentsUntil(currenteditor).length === d) {
						switch(tagName) {
							case 'div':
									if (typeof(tag.attr("class")) !== "undefined") {
										var col = tag.attr("class").split(" ");
										var content = tag.html();

										for (var j = 0; j < col.length; j++) {
											var temp = col[j].split("-");

											if (temp.length === 3 && temp[0] === "col") {
												var config = $("[egaas-class='col']").find(".configuration")[0].outerHTML;
												$(tag).replaceWith(templateDIV(temp[1], temp[2], content, config));
											}
										}
									}
								break;
							case 'p':

								break;
							case 'span':

								break;
							case 'strong':

								break;
						}
					}
				}
			}
			
			$(".demo .lyrow .drag").removeClass("column");
			
			initContainer();
		}
		
		$("#editorModal").modal('hide');
	});
	$("#clear").click(function(e) {
		e.preventDefault();
		clearDemo();
	});
	$('#undo').click(function(){
		stopsave++;
		
		if (undoLayout()) {
			initContainer();
		}
		
		stopsave--;
	});
	$('#redo').click(function(){
		stopsave++;
		
		if (redoLayout()) {
			initContainer();
		}
		
		stopsave--;
	});
	
	var handleSaveLayoutInterval = setInterval(function() {
		if ($("#eGaaSEditor").length) {
			handleSaveLayout();
		} else {
			clearInterval(handleSaveLayoutInterval);
		}
	}, timerSave);
	
	initContainer();
	removeElm();
	
	/* Temp */
	
	$("#getTableType").on('change', function() {
		var type = $(this).val();
		
		if (type === '0') {
			$("#getTableData").hide();
		} else {
			$("#getTableData").show();
		}
	});
	$("#AddColumn").on('click', function() {
		var arr = $(this).prev().clone();
		arr.find(".form-control").val("");
		arr.insertBefore($(this));
	});
}

function resizeCanvas(size) {
	"use strict";
	
	var containerID = document.getElementsByClassName("changeDimension");
	
	if (size === "lg") {
		$(containerID).attr('id', "LG");
	}
	if (size === "md") {
		$(containerID).attr('id', "MD");
	}
	if (size === "sm") {
		$(containerID).attr('id', "SM");
	}
	if (size === "xs") {
		$(containerID).attr('id', "XS");
	}
}

initGenerator();