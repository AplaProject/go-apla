var interval;

function TooltipEllipsis(element) {
	'use strict';
	
	$(window).on('resize load', function(){
		if (el.find(".tooltipEllipsis").length) {
			return false;
		}
	});
	
	var el = $("." + element);
	
	el.each(function() {
		var w = $(this).width();
		var text = $(this).text();
		var l = parseInt($(this).css("padding-left"));
		var t = parseInt($(this).css("padding-top"));
		if (w >= 75) {
			$(this).html('<div class="tooltipEllipsis" style="padding:' + t + 'px ' + l + 'px; margin:-' + t + 'px -' + l + 'px">' + text + '</div>');
		}
	});
	
	if (el) {
		$(".tooltipEllipsisView").remove();
		$("body").append('<div class="tooltipEllipsisView"></div>');
		el.find(".tooltipEllipsis").on('mouseenter', function() {
			var elem = $(this);
			var text = elem.text();
			clearInterval(interval);
			$(".tooltipEllipsis").removeClass("open");
			$(".tooltipEllipsisView").removeClass("open");
			elem.addClass("open");
			TooltipEllipsisView(elem, text);
		});
		el.find(".tooltipEllipsis").on('mouseleave', function() {
			TooltipEllipsisHide();
		});
	}
}
function TooltipEllipsisView(elem, text) {
	'use strict';
	
	var container = $(".tooltipEllipsisView");
	var l = elem.offset().left;
	var t = elem.offset().top + elem.height() + (parseInt(elem.css("padding-top")) * 2);
	
	container.css({"left":l + "px", "top":t + "px"}).html(text).addClass("open");
	
	container.on('mouseenter', function() {
		clearInterval(interval);
	});
	container.on('mouseleave', function() {
		TooltipEllipsisHide();
	});
}
function TooltipEllipsisHide() {
	'use strict';
	
	clearInterval(interval);
	interval = setInterval(function() {
		$(".tooltipEllipsis").removeClass("open");
		$(".tooltipEllipsisView").removeClass("open");
		$(".tooltipEllipsisView").css({"left":"-10000px", "top":"-10000px"});
	}, 10);
}