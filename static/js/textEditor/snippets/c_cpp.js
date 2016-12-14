define("ace/snippets/c_cpp",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "## Contracts Collections\n\
# contract \n\
snippet contract\n\
	contract ${1:YOURCONTRACTNAME} {\n\
		tx {\n\
			${2:/*variable*/}    ${3:/*type*/}    ${4:/*value*/}\n\
		}\n\
		\n\
		func front {\n\
			if ${5:/*condition*/} {\n\
				error ${6:/*text*/}\n\
			}\n\
		}\n\
		\n\
		func main {\n\
			${7:/*variables*/}\n\
			\n\
			${8:DBTransfer(${9:/*code*/})}\n\
			${10:DBInsert(${11:/*code*/})}\n\
			${12:DBUpdate(${13:/*code*/})}\n\
		}\n\
	}\n\
\n\
# tx \n\
snippet tx\n\
	tx {\n\
		${1:/*variable*/}    ${2:/*type*/}    ${3:/*value*/}\n\
	}\n\
\n\
# func front \n\
snippet func front, front\n\
	func front {\n\
		if ${1:/*condition*/} {\n\
			error ${2:/*text*/}\n\
		}\n\
	}\n\
\n\
# func main \n\
snippet func main, main\n\
	func main {\n\
		${1:/*variables*/}\n\
		\n\
		${2:DBTransfer(${3:/*code*/})}\n\
		${4:DBInsert(${5:/*code*/})}\n\
		${6:DBUpdate(${7:/*code*/})}\n\
	}\n\
\n\
";
exports.scope = "c_cpp";

});
