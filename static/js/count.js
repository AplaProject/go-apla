var $scroller = $(window);
var inViewFlagClass = "countUp-view";

function countUp() {
	'use strict';
	
	$("[data-count]").each(countUpInit);
}

function countUpInit(index, element) {
	'use strict';
	
	var $element = $(element);
	$element.css({"width":$element.data("countWidth") + "px", "height":$element.data("countWidth") + "px", "border-width":$element.data("countThickness") + "px", "border-color":$element.data("countOutline")});
	
	$scroller.scroll(function() {
		checkCountUpInVIew($element, index, element);
	});
	checkCountUpInVIew($element, index, element);
}

function checkCountUpInVIew(element, index, elem) {
	'use strict';
	
	var offset = -20;
	
	if (!element.hasClass(inViewFlagClass) && $.Utils.isInView(element, {topoffset: offset})) {
		startCountUp(element, index, elem);
	}
}

function startCountUp(element, index, elem) {
	'use strict';
	
	var container = elem.parentNode;
	var id = "countUp_" + index;
	var countNumber = elem.getAttribute('data-count-number');
	var countPercentage = elem.getAttribute('data-count-percentage');
	var countFont = elem.getAttribute('data-count-font');
	var countFontColor = elem.getAttribute('data-count-font-color');
	var countColor = elem.getAttribute('data-count-color');
	var countOutline = elem.getAttribute('data-count-outline');
	var countFill = elem.getAttribute('data-count-fill');
	var countPie = elem.getAttribute('data-count-pie');
	var countWidth = elem.getAttribute('data-count-width');
	var countThickness = elem.getAttribute('data-count-thickness');
	var countSpeed = elem.getAttribute('data-count-speed');
	var countSeparator = elem.getAttribute('data-count-separator');
	var countDecimal = elem.getAttribute('data-count-decimal');
	var countDecimals = elem.getAttribute('data-count-decimals');
	var countPrefix = elem.getAttribute('data-count-prefix');
	var countSuffix = elem.getAttribute('data-count-suffix');
	var countSpeedPercentage = countSpeed * ((100 / (countPercentage / 100)) / 100);
	
	var dom = document.createElement("div");
	var circuit = document.createElement("div");
	var span1 = document.createElement("span");
	var span2 = document.createElement("span");
	var outline_left = document.createElement("div");
	var outline_right = document.createElement("div");
	var spinner = document.createElement("div");
	var filler = document.createElement("div");
	var mask = document.createElement("div");
	var digit = document.createElement("div");
	
	circuit.className = "circuit";
	circuit.style.backgroundColor = countFill;
	//circuit.setAttribute("style", "background-color:" + countFill);
	outline_left.className = "outline_left";
	outline_right.className = "outline_right";
	span1.style.border = countThickness + "px solid " + countOutline;
	span2.style.border = countThickness + "px solid " + countOutline;
	outline_left.appendChild(span1);
	outline_right.appendChild(span2);
	
	spinner.className = "pie spinner";
	filler.className = "pie filler";
	mask.className = "mask";
	spinner.style.border = countThickness + "px solid " + countColor;
	spinner.style.backgroundColor = countPie;
	filler.style.border = countThickness + "px solid " + countColor;
	filler.style.backgroundColor = countPie;
	
	digit.setAttribute("id", id);
	digit.style.fontSize = countFont;
	digit.style.color = countFontColor;
	digit.className = "digit";
	
	dom.className = "countUp";
	dom.style.width = countWidth + "px";
	dom.style.height = countWidth + "px";
	
	circuit.appendChild(outline_left);
	circuit.appendChild(outline_right);
	circuit.appendChild(spinner);
	circuit.appendChild(filler);
	circuit.appendChild(mask);
	dom.appendChild(circuit);
	dom.appendChild(digit);
	container.replaceChild(dom, elem);
	
	spinner.style.animation = "countUpRotate " + countSpeedPercentage + "s linear forwards";
	filler.style.animation = "countUpFiller " + countSpeedPercentage + "s steps(1, end) forwards";
	mask.style.animation = "countUpMask " + countSpeedPercentage + "s steps(1, end) forwards";
	countUpDigit(id, countNumber, countSpeed, countSeparator, countDecimal, countDecimals, countPrefix, countSuffix, dom, countPercentage);
	
	element.addClass(inViewFlagClass);
}

function countUpDigit(id, countNumber, countSpeed, countSeparator, countDecimal, countDecimals, countPrefix, countSuffix, dom, countPercentage) {
	'use strict';
	
	var countUpOptions = {
		useEasing : true,
		useGrouping : true,
		separator : countSeparator,
		decimal : countDecimal,
		prefix : countPrefix,
		suffix : countSuffix
	};
	
	var countUp = new CountUp(id, 0, countNumber, countDecimals, countSpeed, countUpOptions);
	countUp.start(function() {
		if (countPercentage < 100) {
			dom.className += " stop";
		}
	});
}