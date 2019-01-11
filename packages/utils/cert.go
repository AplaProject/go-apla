// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

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
