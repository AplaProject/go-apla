%yyc c
%yyn c = l.Next()
%yym l.Mark()

%{
package lexer

import (
	"github.com/GenesisKernel/go-genesis/packages/simvolio/token"

	"github.com/cznic/golex/lex"
)

func (l *lexer) scan() lex.Char {
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

[ \t\r]+		// ignore all whitespace

\n				return l.char(token.NewLine)

=				return l.char(token.Assign)
\+				return l.char(token.Plus)
-				return l.char(token.Minus)
\*				return l.char(token.Asterisk)
\/				return l.char(token.Slash)
%				return l.char(token.Percent)
!				return l.char(token.Bang)
\<=				return l.char(token.LtEq)
>=				return l.char(token.GtEq)
\<				return l.char(token.Lt)
>				return l.char(token.Gt)
==				return l.char(token.Eq)
!=				return l.char(token.NotEq)
&&				return l.char(token.And)
\|\|			return l.char(token.Or)

\.\.\. 			return l.char(token.Tail)
\.				return l.char(token.Dot)
,				return l.char(token.Comma)
\(				return l.char(token.LParen)
\)				return l.char(token.RParen)
\{				return l.char(token.LBrace)
\}				return l.char(token.RBrace)
\[				return l.char(token.LBraket)
\]				return l.char(token.RBraket)
:				return l.char(token.Colon)

contract		return l.char(token.Contract)
data			return l.char(token.Data)
condition		return l.char(token.Condition)
action			return l.char(token.Action)
func			return l.char(token.Func)
var				return l.char(token.Var)
true			return l.char(token.True)
false			return l.char(token.False)
if				return l.char(token.If)
else			return l.char(token.Else)
while			return l.char(token.While)
break			return l.char(token.Break)
continue		return l.char(token.Continue)
info			return l.char(token.Info)
warning			return l.char(token.Warning)
error			return l.char(token.Error)
nil				return l.char(token.Nil)
return			return l.char(token.Return)

bool			return l.char(token.TypeBool)
int				return l.char(token.TypeInt)
float			return l.char(token.TypeFloat)
money			return l.char(token.TypeMoney)
string			return l.char(token.TypeString)
bytes			return l.char(token.TypeBytes)
array			return l.char(token.TypeArray)
map				return l.char(token.TypeMap)
file			return l.char(token.TypeFile)

\${identifier}	return l.char(token.ExtendVar)
{identifier}	return l.char(token.Ident)
{int}			return l.char(token.Int)
{float}			return l.char(token.Float)
{string}		return l.char(token.String)

%%
	if c, ok := l.Abort(); ok {
		if c == -1 {
			return l.char(token.EOF)
		}
		return l.char(token.TokenType(c))
	}

	goto yyAction
}