var $scroller;
var inViewFlagClass;

function countUp() {
	$("[data-count]").each(function(index, element) {
		var container = element.parentNode;
		var id = "countUp_" + index;
		var countNumber = element.getAttribute('data-count-number');
		var countPercentage = element.getAttribute('data-count-percentage');
		var countFont = element.getAttribute('data-count-font');
		var countColor = element.getAttribute('data-count-color');
		var countOutline = element.getAttribute('data-count-outline');
		var countWidth = element.getAttribute('data-count-width');
		var countHeight = element.getAttribute('data-count-height');
		var countThickness = element.getAttribute('data-count-thickness');
		var countDiameter = (element.getAttribute('data-count-diameter') - countThickness) / 2;
		var countSpeed = element.getAttribute('data-count-speed');
		var countSeparator = element.getAttribute('data-count-separator');
		var countDecimal = element.getAttribute('data-count-decimal');
		var countPrefix = element.getAttribute('data-count-prefix');
		var countSuffix = element.getAttribute('data-count-suffix');
		
		var dom = document.createElement("div");
		var canvas = document.createElement("canvas");
		var span = document.createElement("span");
		canvas.setAttribute("data-classyloader", "");
		canvas.setAttribute("data-trigger-in-view", "true");
		canvas.setAttribute("data-percentage", countPercentage);
		canvas.setAttribute("data-speed", countSpeed);
		canvas.setAttribute("data-font-size", countFont);
		canvas.setAttribute("data-width", countWidth);
		canvas.setAttribute("data-height", countHeight);
		canvas.setAttribute("data-diameter", countDiameter);
		canvas.setAttribute("data-line-color", countColor);
		canvas.setAttribute("data-remaining-line-color", countOutline);
		canvas.setAttribute("data-line-width", countThickness);
		canvas.setAttribute("data-show-text", "false");
		span.setAttribute("id", id);
		span.setAttribute("style", "font-size:" + countFont);
		dom.className = "countUp";
		dom.appendChild(canvas);
		dom.appendChild(span);
		container.replaceChild(dom, element);
		
		countUpStart(id, countNumber, countSpeed, countSeparator, countDecimal, countPrefix, countSuffix);
	});
	
	countStart();
}

function countUpStart(id, countNumber, countSpeed, countSeparator, countDecimal, countPrefix, countSuffix) {
	var countUpOptions = {
		useEasing : true,
		useGrouping : true,
		separator : countSeparator,
		decimal : countDecimal,
		prefix : countPrefix,
		suffix : countSuffix
	};
	
	var countUp = new CountUp(id, 0, countNumber, 0, countSpeed, countUpOptions);
	countUp.start();
}

function countStart() {
	$scroller = $(window);
	inViewFlagClass = "js-is-in-view";
	
	$("[data-classyloader]").each(initClassyLoader);
}

function initClassyLoader() {
	var $element = $(this);
	var options  = $element.data();
	
	if(options) {
		if(options.triggerInView) {
			$scroller.scroll(function() {
				checkLoaderInVIew($element, options);
			});
			checkLoaderInVIew($element, options);
		} else {
			startLoader($element, options);
		}
	}
}

function checkLoaderInVIew(element, options) {
	var offset = -20;
	
	if(!element.hasClass(inViewFlagClass) && $.Utils.isInView(element, {topoffset: offset})) {
		startLoader(element, options);
	}
}

function startLoader(element, options) {
	element.ClassyLoader(options).addClass(inViewFlagClass);
}