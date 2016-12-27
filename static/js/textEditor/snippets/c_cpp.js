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
# func Balance \n\
snippet Balance()\n\
	Balance(${1:/*code*/})\n\
\n\
# func Money \n\
snippet Money()\n\
	Money(${1:/*code*/})\n\
\n\
# func StateParam \n\
snippet StateParam()\n\
	StateParam(${1:/*code*/})\n\
\n\
# func AddressToId \n\
snippet AddressToId()\n\
	AddressToId(${1:/*code*/})\n\
\n\
# func DBTransfer \n\
snippet DBTransfer()\n\
	DBTransfer(${1:/*code*/})\n\
\n\
# func DBInsert \n\
snippet DBInsert()\n\
	DBInsert(${1:/*code*/})\n\
\n\
# func DBUpdate \n\
snippet DBUpdate()\n\
	DBUpdate(${1:/*code*/})\n\
\n\
# func DBAmount \n\
snippet DBAmount()\n\
	DBAmount(${1:/*code*/})\n\
\n\
# func DBInt \n\
snippet DBInt()\n\
	DBInt(${1:/*code*/})\n\
\n\
# func DBIntExt \n\
snippet DBIntExt()\n\
	DBIntExt(${1:/*code*/})\n\
\n\
# func DBIntWhere \n\
snippet DBIntWhere()\n\
	DBIntWhere(${1:/*code*/})\n\
\n\
# func DBGetList \n\
snippet DBGetList()\n\
	DBGetList(${1:/*code*/})\n\
\n\
# func TableTx \n\
snippet TableTx()\n\
	TableTx(${1:/*code*/})\n\
\n\
# func Table \n\
snippet Table()\n\
	Table(${1:/*code*/})\n\
\n\
# func Int \n\
snippet Int()\n\
	Int(${1:/*code*/})\n\
\n\
# func StateValue \n\
snippet StateValue()\n\
	StateValue(${1:/*code*/})\n\
\n\
# func Println \n\
snippet Println()\n\
	Println(${1:/*code*/})\n\
\n\
";
exports.scope = "c_cpp";

});
