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

package smart

import (
	"math"
	"strconv"
)

func parseFloat(x interface{}) (float64, error) {
	var (
		fx  float64
		err error
	)
	switch v := x.(type) {
	case float64:
		fx = v
	case int64:
		fx = float64(v)
	case string:
		if fx, err = strconv.ParseFloat(v, 64); err != nil {
			return 0, errFloat
		}
	default:
		return 0, errFloat
	}
	return fx, nil
}

func isValidFloat(x float64) bool {
	return !(math.IsNaN(x) || math.IsInf(x, 1) || math.IsInf(x, -1))
}

// Floor returns the greatest integer value less than or equal to x
func Floor(x interface{}) (int64, error) {
	fx, err := parseFloat(x)
	if err != nil {
		return 0, err
	}
	if fx = math.Floor(fx); isValidFloat(fx) {
		return int64(fx), nil
	}
	return 0, errFloatResult
}

// Log returns the natural logarithm of x
func Log(x interface{}) (float64, error) {
	fx, err := parseFloat(x)
	if err != nil {
		return 0, err
	}
	if fx = math.Log(fx); isValidFloat(fx) {
		return fx, nil
	}
	return 0, errFloatResult
}

// Log10 returns the decimal logarithm of x
func Log10(x interface{}) (float64, error) {
	fx, err := parseFloat(x)
	if err != nil {
		return 0, err
	}
	if fx = math.Log10(fx); isValidFloat(fx) {
		return fx, nil
	}
	return 0, errFloatResult
}

// Pow returns x**y, the base-x exponential of y
func Pow(x, y interface{}) (float64, error) {
	fx, err := parseFloat(x)
	if err != nil {
		return 0, err
	}
	fy, err := parseFloat(y)
	if err != nil {
		return 0, err
	}
	if fx = math.Pow(fx, fy); isValidFloat(fx) {
		return fx, nil
	}
	return 0, errFloatResult
}

// Round returns the nearest integer, rounding half away from zero
func Round(x interface{}) (int64, error) {
	fx, err := parseFloat(x)
	if err != nil {
		return 0, err
	}
	if fx = math.Round(fx); isValidFloat(fx) {
		return int64(fx), nil
	}
	return 0, errFloatResult
}

// Sqrt returns the square root of x
func Sqrt(x interface{}) (float64, error) {
	fx, err := parseFloat(x)
	if err != nil {
		return 0, err
	}
	if fx = math.Sqrt(fx); isValidFloat(fx) {
		return fx, nil
	}
	return 0, errFloatResult
}
