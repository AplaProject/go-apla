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

package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscape(t *testing.T) {
	var cases = []struct {
		Source   string
		Expected template.HTML
	}{
		{`'test'`, `''test''`},
		{"`test`", "` + \"`\" + `test` + \"`\" + `"},
		{`100%`, `100%%`},
	}

	for _, v := range cases {
		assert.Equal(t, v.Expected, escape(v.Source))
	}
}

func tempContract(appID int, conditions, value string) (string, error) {
	file, err := ioutil.TempFile("", "contract")
	if err != nil {
		return "", err
	}
	defer file.Close()

	file.Write([]byte(fmt.Sprintf(`// +prop AppID = %d
// +prop Conditions = '%s'
%s`, appID, conditions, value)))

	return file.Name(), nil
}

func TestLoadSource(t *testing.T) {
	value := "contract Test {}"

	path, err := tempContract(5, "true", value)
	assert.NoError(t, err)

	source, err := loadSource(path)
	assert.NoError(t, err)

	assert.Equal(t, &contract{
		Name:       filepath.Base(path),
		Source:     template.HTML(value + "\n"),
		Conditions: template.HTML("true"),
		AppID:      5,
	}, source)
}
