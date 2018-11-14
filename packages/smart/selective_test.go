package smart

import (
	"testing"
)

func concat(raw, wrapper string) string {
	return wrapper + raw + wrapper
}

func BenchmarkWrappingConcat(b *testing.B) {
	concat("someRawString", `"`)
}

func BenchmarkWrapWithBuffer(b *testing.B) {
	wrapString("someRawString", `"`)
}
