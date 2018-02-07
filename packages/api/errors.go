//MIT License
//
//Copyright (c) 2016 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.
package api

var (
	apiErrors = map[string]string{
		`E_CONTRACT`:      `There is not %s contract`,
		`E_DBNIL`:         `DB is nil`,
		`E_ECOSYSTEM`:     `Ecosystem %d doesn't exist`,
		`E_EMPTYPUBLIC`:   `Public key is undefined`,
		`E_EMPTYSIGN`:     `Signature is undefined`,
		`E_HASHWRONG`:     `Hash is incorrect`,
		`E_HASHNOTFOUND`:  `Hash has not been found`,
		`E_HEAVYPAGE`:     `This page is heavy`,
		`E_INSTALLED`:     `Apla is already installed`,
		`E_INVALIDWALLET`: `Wallet %s is not valid`,
		`E_NOTFOUND`:      `Page not found`,
		`E_NOTINSTALLED`:  `Apla is not installed`,
		`E_PERMISSION`:    `Permission denied`,
		`E_QUERY`:         `DB query is wrong`,
		`E_RECOVERED`:     `API recovered`,
		`E_REFRESHTOKEN`:  `Refresh token is not valid`,
		`E_SERVER`:        `Server error`,
		`E_SIGNATURE`:     `Signature is incorrect`,
		`E_UNKNOWNSIGN`:   `Unknown signature`,
		`E_STATELOGIN`:    `%s is not a membership of ecosystem %s`,
		`E_TABLENOTFOUND`: `Table %s has not been found`,
		`E_TOKEN`:         `Token is not valid`,
		`E_TOKENEXPIRED`:  `Token is expired by %s`,
		`E_UNAUTHORIZED`:  `Unauthorized`,
		`E_UNDEFINEVAL`:   `Value %s is undefined`,
		`E_UNKNOWNUID`:    `Unknown uid`,
		`E_VDE`:           `Virtual Dedicated Ecosystem %d doesn't exist`,
		`E_VDECREATED`:    `Virtual Dedicated Ecosystem is already created`,
	}
)
