package tags

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_StylesheetTag(t *testing.T) {
	r := require.New(t)
	tcases := []struct {
		Options  Options
		Expected string
	}{
		{Options: Options{"href": "styles.css"}, Expected: `<link href="styles.css" rel="stylesheet" />`},
		{Options: Options{}, Expected: `<link rel="stylesheet" />`},
		{Options: Options{"style": "text/scss"}, Expected: `<link rel="stylesheet" style="text/scss" />`},
	}

	for _, tcase := range tcases {
		tag := StylesheetTag(tcase.Options)
		r.Equal(tcase.Expected, tag.String())
	}
}

func Test_JavascriptTag(t *testing.T) {
	r := require.New(t)
	tcases := []struct {
		Options  Options
		Expected string
	}{
		{Options: Options{"src": "script.js"}, Expected: `<script src="script.js"></script>`},
		{Options: Options{"src": "script.js", "body": `alert("hello")`}, Expected: `<script src="script.js"></script>`},
		{Options: Options{"body": `alert("hello")`}, Expected: `<script>alert("hello")</script>`},
	}

	for _, tcase := range tcases {
		tag := JavascriptTag(tcase.Options)
		r.Equal(tcase.Expected, tag.String())
	}
}
