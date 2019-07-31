// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package script

import "errors"

const (
	eContractLoop    = `there is loop in %s contract`
	eSysVar          = `system variable $%s cannot be changed`
	eTypeParam       = `parameter %d has wrong type`
	eUndefinedParam  = `%s is not defined`
	eUnknownContract = `unknown contract %s`
	eWrongParams     = `function %s must have %d parameters`
	eArrIndex        = `index of array cannot be type %s`
	eMapIndex        = `index of map cannot be type %s`
	eUnknownIdent    = `unknown identifier %s`
	eWrongVar        = `wrong var %v`
	eDataType        = `expecting type of the data field [Ln:%d Col:%d]`
	eDataName        = `expecting name of the data field [Ln:%d Col:%d]`
	eDataTag         = `unexpected tag [Ln:%d Col:%d]`
)

var (
	errContractPars    = errors.New(`wrong contract parameters`)
	errWrongCountPars  = errors.New(`wrong count of parameters`)
	errDivZero         = errors.New(`divided by zero`)
	errUnsupportedType = errors.New(`unsupported combination of types in the operator`)
	errMaxArrayIndex   = errors.New(`The index is out of range`)
	errMaxMapCount     = errors.New(`The maxumim length of map`)
	errRecursion       = errors.New(`The contract can't call itself recursively`)
	errUnclosedArray   = errors.New(`unclosed array initialization`)
	errUnclosedMap     = errors.New(`unclosed map initialization`)
	errUnexpKey        = errors.New(`unexpected lexem; expecting string key`)
	errUnexpColon      = errors.New(`unexpected lexem; expecting colon`)
	errUnexpComma      = errors.New(`unexpected lexem; expecting comma`)
	errUnexpValue      = errors.New(`unexpected lexem; expecting string, int value or variable`)
	errCondWrite       = errors.New(`'conditions' cannot call contracts or functions which can modify the blockchain database.`)
	errMultiIndex      = errors.New(`multi-index is not supported`)
	errSelfAssignment  = errors.New(`self assignment`)
	errEndExp          = errors.New(`unexpected end of the expression`)
	errOper            = errors.New(`unexpected operator; expecting operand`)
)
