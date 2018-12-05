%yyc c
%yyn c = l.Next()
%yym l.Mark()

%{
package parser

import (
	"strconv"

	"github.com/cznic/golex/lex"
)

func (l *lexer) scan(lval *yySymType) lex.Char {
	c := l.Enter()
%}

digit			[0-9]|{unicodeDigit}
identifier		{letter}({letter}|{digit})*
int				[0-9]+
float			{int}\.[0-9]*
string			\"({letter}|{digit})*\"
letter			[_a-zA-Z]|{unicodeLetter}
unicodeDigit	\x81
unicodeLetter	\x80

%%

	c = l.Rule0()

[ \t\r\n ]+		// ignore all whitespace

\+=				return l.char(ADD_ASSIGN)
-=				return l.char(SUB_ASSIGN)
\*=				return l.char(MUL_ASSIGN)
\/=				return l.char(DIV_ASSIGN)
%=				return l.char(MOD_ASSIGN)
=				return l.char(ASSIGN)

\+\+			return l.char(INC)
--				return l.char(DEC)

&&				return l.char(AND)
\|\|			return l.char(OR)

\+				return l.char(ADD)
-				return l.char(SUB)
\*				return l.char(MUL)
\/				return l.char(DIV)
%				return l.char(MOD)

==				return l.char(EQ)
!=				return l.char(NOT_EQ)
!				return l.char(NOT)

\<=				return l.char(LTE)
>=				return l.char(GTE)
\<				return l.char(LT)
>				return l.char(GT)

\.\.\. 			return l.char(ELLIPSIS)

\.				return l.char(DOT)
,				return l.char(COMMA)
:				return l.char(COLON)
\(				return l.char(LPAREN)
\)				return l.char(RPAREN)
\{				return l.char(LBRACE)
\}				return l.char(RBRACE)
\[				return l.char(LBRAKET)
\]				return l.char(RBRAKET)

contract		return l.char(CONTRACT)
data			return l.char(DATA)
condition		return l.char(CONDITION)
action			return l.char(ACTION)
func			return l.char(FUNC)
var				return l.char(VAR)
true			return l.char(TRUE)
false			return l.char(FALSE)
if				return l.char(IF)
else			return l.char(ELSE)
while			return l.char(WHILE)
break			return l.char(BREAK)
continue		return l.char(CONTINUE)
info			return l.char(INFO)
warning			return l.char(WARNING)
error			return l.char(ERROR)
nil				return l.char(NIL)
return			return l.char(RETURN)

bool			return l.char(T_BOOL)
int				return l.char(T_INT)
float			return l.char(T_FLOAT)
money			return l.char(T_MONEY)
string			return l.char(T_STRING)
bytes			return l.char(T_BYTES)
array			return l.char(T_ARRAY)
map				return l.char(T_MAP)
file			return l.char(T_FILE)

\${identifier}	{
					lval.s = string(l.TokenBytes(nil))
					return l.char(EXTEND_VAR)
				}
{int}			{
					lval.i, _ = strconv.Atoi(string(l.TokenBytes(nil)))
					return l.char(INT)
				}
{float}			{
					lval.f, _ = strconv.ParseFloat(string(l.TokenBytes(nil)), 64)
					return l.char(FLOAT)
				}
{identifier}	{
					lval.s = string(l.TokenBytes(nil))
					return l.char(IDENT)
				}
{string}		{
					lval.s = string(l.TokenBytes(nil))
					return l.char(STRING)
				}

%%
	if c, ok := l.Abort(); ok {
		return l.char(c)
	}

	goto yyAction
}