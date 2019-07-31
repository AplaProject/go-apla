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

package utils

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
)

var (
	errParseCert     = errors.New("Failed to parse certificate")
	errParseRootCert = errors.New("Failed to parse root certificate")
)

type Cert struct {
	cert *x509.Certificate
}

func (c *Cert) Validate(pem []byte) error {
	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM(pem); !ok {
		return errParseRootCert
	}

	if _, err := c.cert.Verify(x509.VerifyOptions{Roots: roots}); err != nil {
		return err
	}

	return nil
}

func (c *Cert) EqualBytes(bs ...[]byte) bool {
	for _, b := range bs {
		other, err := parseCert(b)
		if err != nil {
			return false
		}

		if c.cert.Equal(other) {
			return true
		}
	}

	return false
}

func parseCert(b []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errParseCert
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func ParseCert(b []byte) (c *Cert, err error) {
	cert, err := parseCert(b)
	if err != nil {
		return nil, err
	}

	return &Cert{cert}, nil
}
