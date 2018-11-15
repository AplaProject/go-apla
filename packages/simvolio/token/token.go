package token

// https://github.com/mmyoji/go-monkey

//go:generate stringer -type=TokenType

type TokenType int

const (
	EOF TokenType = iota

	Illegal // Illegal token

	NewLine // \n

	// Identifiers + literals
	Ident  // add, foobar, x, y, ...
	Int    // 314
	Float  // 3.14
	String // "string"

	// Operators
	Assign   // =
	Plus     // +
	Minus    // -
	Asterisk // *
	Slash    // /
	Percent  // %
	Bang     // !
	Lt       // <
	Gt       // >
	LtEq     // <=
	GtEq     // >=
	Eq       // ==
	NotEq    // !=
	And      // &&
	Or       // ||

	// Delimiters
	Dot     // .
	Comma   // ,
	LParen  // (
	RParen  // )
	LBrace  // {
	RBrace  // }
	LBraket // [
	RBraket // ]
	Colon   // :
	Tail    // ...

	// Keywords
	Contract  // contract
	Data      // data
	Condition // condition
	Action    // action
	Func      // func
	Var       // var
	ExtendVar // $foo
	True      // true
	False     // false
	If        // if
	Else      // else
	While     // while
	Break     // break
	Continue  // continue
	Info      // info
	Warning   // warning
	Error     // error
	Nil       // nil
	Return    // return

	// Types
	TypeBool   // bool
	TypeInt    // int
	TypeFloat  // float
	TypeMoney  // money
	TypeString // string
	TypeBytes  // bytes
	TypeArray  // array
	TypeMap    // map
	TypeFile   // file
)
